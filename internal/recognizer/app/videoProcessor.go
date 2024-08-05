package recognizer

import (
	"context"
	"errors"
	"fmt"
	"image/color"
	"log"
	"math"
	"os"
	"path"
	"sync"
	"sync/atomic"

	"gocv.io/x/gocv"

	face "go_cv_test/internal/recognizer"
)

// Путь до папки с моделями. Папка должна содержать следующий файлы: dlib_face_recognition_resnet_model_v1.dat,
// mmod_human_face_detector.dat, shape_predictor_68_face_landmarks.dat. Архивы с этими файлами можно скачать из
// https://github.com/davisking/dlib-models.
const modelsPath = "./internal/recognizer/app/models"

// Путь до папки с персонами. Папка должна содержать на первом уровне папки, где название папки ― имя персоны,
// а на втором уровне ― файлы с фотографиями лица соответствующей персоны.
const personsPath = "./internal/recognizer/app/persons"

// ID оборудования для получения видеопотока. В нашем случае 0 ― это ID стандартной веб-камеры.
const deviceID = 0

// Параметры векторизации, которые влияют на качество получаемого вектора:
const (
	padding   = 0.2 // насколько увеличивать квадрат выявленного лица;
	jittering = 30  // кол-во генерируемых немного сдвинутых и повёрнутых копий лица.
)

// Синий цвет.
var blue = color.RGBA{
	R: 0,
	G: 0,
	B: 255,
	A: 0,
}

// Минимальное расстояние соответствия персоне.
const matchDistance = 0.5

// Структура, описывающая персону.
type Person struct {
	// Имя персоны.
	Name string

	// Список дескрипторов лица персоны.
	Descriptors []face.Descriptor
}

type VideoStatus int

const (
	InQueue VideoStatus = 0

	Processing VideoStatus = 1

	Error VideoStatus = 2

	Canceled VideoStatus = 3

	Successful VideoStatus = 4

	Paused VideoStatus = 5
)

type Video struct {
	Id         int32       `json:"id"`
	Status     VideoStatus `json:"video_status"`
	Percentage float64     `json:"percentage"`
	Name       string      `json:"name"`
}

type VideoProcessor struct {
	//stores max amount of CPUs
	CPUs int
	//this chanel is used for limiting max amount of working goroutines
	chanel chan struct{}
	//used for counting amount of goroutines
	grCounter atomic.Int32
	//used for generating id for each processing video
	processId atomic.Int32
	//stores videos
	videos map[int32]Video
	//used for pausing goroutines with provived video id
	switcher chan int32
	//used for runVideoUpdater(), this is a buffered channel
	dataBuffer chan Video
}

// accepts video id and returns founded video
func (vP *VideoProcessor) GetVideo(id int32) (Video, error) {
	val, ok := vP.videos[id]
	if !ok {
		return Video{}, fmt.Errorf("unable to find video with given id")
	}
	return val, nil
}

// switches state for video with provided id
func (vP *VideoProcessor) SwitchState(id int32) error {
	_, err := vP.GetVideo(id)
	if err != nil {
		return errors.New("unable to locate video with providen id")
	}
	vP.switcher <- id
	return nil
}

// returns ready for work VideoProcessor,
func GetVideoProcessor(numOfCores int) *VideoProcessor {
	if numOfCores < 1 {
		numOfCores = 1
	}
	vp := VideoProcessor{
		CPUs:       numOfCores,
		chanel:     make(chan struct{}, numOfCores),
		videos:     make(map[int32]Video),
		switcher:   make(chan int32),
		dataBuffer: make(chan Video, numOfCores)}
	vp.runVideoUpdater()
	return &vp
}

// This method is used for running goroutine that can write everything that comes from channel to a map, we update video state here
func (vP *VideoProcessor) runVideoUpdater() {
	go func() {
		for {
			select {
			case data := <-vP.dataBuffer:
				vP.videos[data.Id] = data
			case id := <-vP.switcher:
				video := vP.videos[id]
				if video.Status == 5 {
					video.Status = 1
					vP.videos[id] = video
				}
				if video.Status == 1 {
					video.Status = 5
					vP.videos[id] = video
				}
			}
		}
	}()
}

// fileName is used for nothing, but logging file name
func (vP *VideoProcessor) RunRecognizer(ctx context.Context, cancel context.CancelCauseFunc, videoFile, fileName string, wg *sync.WaitGroup) {
	var id = vP.processId.Add(1)
	vidInfo := Video{Id: id,
		Status:     0,
		Name:       fileName,
		Percentage: 0.0}
	//here we write data about video into chanel, which is read by one goroutine
	vP.dataBuffer <- vidInfo
	//here we signal that we want to start a new video processing, it will waint until channel will have space
	vP.chanel <- struct{}{}
	//video is in process
	vidInfo.Status = 1
	vP.dataBuffer <- vidInfo

	// Инициализация детектора лиц, который будет выявлять лица.
	detector, err := face.NewDetector(path.Join(modelsPath, "mmod_human_face_detector.dat"))
	// Check that detecor init successful
	if err != nil {
		fmt.Printf("Error at detecor init stage: %s\n", err.Error())
		vidInfo.Status = 2
		vP.dataBuffer <- vidInfo
		cancel(errors.New("unable to init face detector: " + err.Error()))
		wg.Done()
		return
	}
	defer detector.Close()

	// Инициализация распознавателя лиц, который будет векторизовывать лица.
	recognizer, err := face.NewRecognizer(
		path.Join(modelsPath, "shape_predictor_68_face_landmarks.dat"),
		path.Join(modelsPath, "dlib_face_recognition_resnet_model_v1.dat"))
	// Check that recognizer init successful
	if err != nil {
		fmt.Printf("Error at recognizer init stage %s\n", err.Error())
		vidInfo.Status = 2
		vP.dataBuffer <- vidInfo
		cancel(errors.New("unable to init face recognizer: " + err.Error()))
		wg.Done()
		return
	}
	defer recognizer.Close()

	// Инициализация базы персон.
	persons := loadPersons(detector, recognizer, personsPath)

	// Init video capture from file
	video, err := gocv.VideoCaptureFile(videoFile)
	if err != nil {
		fmt.Println(err)
		cancel(errors.New("unable to process file"))
		wg.Done()
		return
	}
	defer video.Close()

	//count goroutine id
	var gr = vP.grCounter.Add(1)
	defer vP.grCounter.Add(-1)

	var frame_counter int64 = 1
	var total_frames = video.Get(gocv.VideoCaptureProperties(7))

	// Инициализация изображения для очередного кадра.
	img := gocv.NewMat()
	defer img.Close()

	fmt.Printf("start reading video from: %s\n", videoFile)
	for {
		var progress = float64(frame_counter) / total_frames * 100
		select {
		//if request was canceled, goroutine is shutting down
		case <-ctx.Done():
			vidInfo.Status = 3
			vidInfo.Percentage = progress
			vP.dataBuffer <- vidInfo
			cancel(errors.New("goroutine was canceled due to context cancel"))
			wg.Done()
			return
		default:
			if vidInfo.Status == 1 {
				vidInfo.Status = 1
				vidInfo.Percentage = progress
				vP.dataBuffer <- vidInfo
				if ok := video.Read(&img); !ok {
					fmt.Printf("cannot read video from file %s\n", videoFile)
					vidInfo.Status = 4
					vidInfo.Percentage = 100.0
					vP.dataBuffer <- vidInfo
					wg.Done()
					return
				}
				if img.Empty() {
					continue
				}
				// Выявляем лица в кадре.
				detects, err := detector.Detect(img)
				if err != nil {
					log.Fatalf("detect faces: %v", err)
				}
				// detect faces
				// Для каждого выявленного лица.
				for _, detect := range detects {

					// Получаем вектор выявленного лица.
					descriptor, err := recognizer.Recognize(img, detect.Rectangle, padding, jittering)
					if err != nil {
						log.Fatalf("recognize face: %v", err)
					}

					// Ищем в массиве векторов известных лиц наиболее близкое (по евклиду) лицо.
					person, distance := findPerson(persons, descriptor)

					// Рисуем прямоугольник выявленного лица.
					//gocv.Rectangle(&img, detect.Rectangle, blue, 1)

					// Если расстояние между найденным известным лицом и выявленным лицом меньше
					// какого-то порога, то пишем имя найденного известного лица над нарисованным
					// прямоугольником.
					if distance <= matchDistance {
						log.Printf("goroutine: %d, processId: %d - %.2f%%: found %s on frame %d of %s\n", gr, id, progress, person.Name, frame_counter, fileName)
						//gocv.PutText(&img, person.Name, image.Point{
						//	X: detect.Rectangle.Min.X,
						//	Y: detect.Rectangle.Min.Y,
						//}, gocv.FontHersheyComplex, 1, blue, 1)
					}
					frame_counter++
				}
			}
		}
	}
}

// Вычисление евклидового расстояния.
func euclidianDistance(a, b face.Descriptor) float64 {
	var sum float64
	for i := range a {
		sum += math.Pow(float64(a[i])-float64(b[i]), 2)
	}
	return math.Sqrt(sum)
}

// Функция поиска наиболее близкой, заданному дескриптору, персоны.
func findPerson(persons []Person, descriptor face.Descriptor) (Person, float64) {
	// Объявляем переменные, которые будут хранить результаты поиска
	var minPerson Person
	var minDistance = math.MaxFloat64

	// Проходимся по каждой персоне.
	for _, person := range persons {
		// Проходимся по каждому дескриптору персоны.
		for _, personDescriptor := range person.Descriptors {
			// Вычисляем расстояние между дескриптором персоны и заданным дескриптором.
			distance := euclidianDistance(personDescriptor, descriptor)

			// Если полученное расстояние меньше текущего минимального, то сохраняем персону и расстояние в
			// переменные результатов.
			if distance < minDistance {
				minDistance = distance
				minPerson = person
			}
		}
	}
	return minPerson, minDistance
}

// Функция загрузки базы персон.
func loadPersons(detector *face.Detector, recognizer *face.Recognizer, personsPath string) (persons []Person) {
	// Читаем директорию, получаем массив его содержимого (информацию о файлах и папках).
	personsDirs, err := os.ReadDir(personsPath)
	//personsDirs, err := ioutil.ReadDir(personsPath)
	if err != nil {
		log.Fatalf("read persons directory: %v", err)
	}

	// По каждому элементу из директории персон.
	for _, personDir := range personsDirs {
		// Пропускаем не директории.
		if !personDir.IsDir() {
			continue
		}

		// Формируем персону.
		person := Person{
			// Имя персоны ― название папки.
			Name: personDir.Name(),
		}

		// Читаем директорию персоны.
		personsFiles, err := os.ReadDir(path.Join(personsPath, personDir.Name()))
		if err != nil {
			log.Fatalf("read person directory: %v", err)
		}

		// По каждому элементу из директории персоны.
		for _, personFile := range personsFiles {
			// Пропускаем если директория.
			if personFile.IsDir() {
				continue
			}

			filePath := path.Join(personsPath, personDir.Name(), personFile.Name())

			// Читаем и декодируем изображение.
			img := gocv.IMRead(filePath, gocv.IMReadUnchanged)

			// Если не удалось прочитать файл и декодировать изображение, то пропускаем файл.
			if img.Empty() {
				continue
			}

			// Выявляем лица на изображении.
			detects, err := detector.Detect(img)
			if err != nil {
				log.Fatalf("detect on person image: %v", err)
			}

			// Если кол-во лиц не 1, то завершаем программу с ошибкой.
			if len(detects) != 1 {
				log.Fatalf("multple faces detected on photo %s", filePath)
			}

			// Получаем вектор лица на изображении.
			descriptor, err := recognizer.Recognize(img, detects[0].Rectangle, padding, jittering)
			if err != nil {
				log.Fatalf("recognize persons face: %v", err)
			}

			// Добавляем вектор в массив векторов персоны.
			person.Descriptors = append(person.Descriptors, descriptor)

			// Освобождаем память, выделенную под изображение.
			err = img.Close()
			if err != nil {
				log.Fatalf("close image: %v", err)
			}
		}

		// Добавляем очередную персону в массив персон.
		persons = append(persons, person)
	}

	return persons
}
