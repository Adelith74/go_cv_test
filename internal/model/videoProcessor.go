package videoProcessor

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"

	"gocv.io/x/gocv"
)

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
	//used for xml classifier file
	XMLfile string
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
func (vP *VideoProcessor) SwitchState(id int32) {
	vP.switcher <- id
}

// returns ready for work VideoProcessor,
func GetVideoProcessor(numOfCores int) *VideoProcessor {
	if numOfCores < 1 {
		numOfCores = 1
	}
	vp := VideoProcessor{
		CPUs:       numOfCores,
		chanel:     make(chan struct{}, numOfCores),
		XMLfile:    "../haarcascade_frontalface_default.xml",
		videos:     make(map[int32]Video),
		switcher:   make(chan int32),
		dataBuffer: make(chan Video, numOfCores)}
	vp.runVideoUpdater()
	return &vp
}

// This method is used for running goroutine that can write everything that comes from channel to a map
func (vP *VideoProcessor) runVideoUpdater() {
	go func() {
		for {
			var data = <-vP.dataBuffer
			vP.videos[data.Id] = data
		}
	}()
}

// fileName is used nothing, but logging file name
func (vP *VideoProcessor) ProcessVideo(ctx context.Context, videoFile, xmlFile, fileName string, wg *sync.WaitGroup) {
	var id = vP.processId.Add(1)
	status := true
	vidInfo := Video{Id: id,
		Status:     0,
		Name:       fileName,
		Percentage: 0.0}
	//here we write data about video into chanel, which is read by one goroutine
	vP.dataBuffer <- vidInfo
	//here we signal that we want to start a new video processing, it will waint until channel will have space
	vP.chanel <- struct{}{}
	vidInfo.Status = 1
	vP.dataBuffer <- vidInfo
	//open video file
	video, err := gocv.VideoCaptureFile(videoFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer video.Close()

	//count goroutine id
	var gr = vP.grCounter.Add(1)
	defer vP.grCounter.Add(-1)

	var frame_counter int64 = 1
	var total_frames = video.Get(gocv.VideoCaptureProperties(7))

	// prepare image matrix
	img := gocv.NewMat()
	defer img.Close()

	// load classifier to recognize faces
	classifier := gocv.NewCascadeClassifier()
	defer classifier.Close()

	if !classifier.Load(xmlFile) {
		fmt.Printf("Error reading cascade file: %v\n", xmlFile)
		vidInfo.Status = 2
		vP.dataBuffer <- vidInfo
		return
	}

	fmt.Printf("start reading video from: %s\n", videoFile)

	//here we look at what status video has, if it's changes, we perform action as a response
	for {
		var progress = float64(frame_counter) / total_frames * 100
		select {
		case sw := <-vP.switcher:
			if sw == id {
				status = !status
			}
			if status {
				log.Printf("goroutine: %d, processId: %d - %.2f%%: Processing of %s was resumed\n", gr, id, progress, fileName)
				vidInfo.Status = 1
				vidInfo.Percentage = progress
				vP.dataBuffer <- vidInfo
			} else {
				log.Printf("goroutine: %d, processId: %d - %.2f%%: Processing of %s was paused\n", gr, id, progress, fileName)
				vidInfo.Status = 5
				vidInfo.Percentage = progress
				vP.dataBuffer <- vidInfo
			}
		//if request was canceled, goroutine is shutting down
		case <-ctx.Done():
			vidInfo.Status = 3
			vidInfo.Percentage = progress
			vP.dataBuffer <- vidInfo
			wg.Done()
			return
		default:
			if status {
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
				// detect faces
				rects := classifier.DetectMultiScale(img)
				log.Printf("goroutine: %d, processId: %d - %.2f%%: found %d faces on frame %d of %s\n", gr, id, progress, len(rects), frame_counter, fileName)
				frame_counter++
			}
		}
	}

}
