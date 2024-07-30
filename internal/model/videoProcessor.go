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
	Status     VideoStatus `json:"video_status"`
	Percentage float64     `json:"percentage"`
	Name       string      `json:"name"`
}

type VideoProcessor struct {
	CPUs      int
	Chanel    chan struct{}
	XMLfile   string
	grCounter atomic.Int32
	processId atomic.Int32
	Videos    map[int32]Video
	Switcher  chan int32
}

func (vP *VideoProcessor) GetVideo(id int32) (Video, error) {
	val, ok := vP.Videos[id]
	if !ok {
		return Video{}, fmt.Errorf("unable to find video with given id")
	}
	return val, nil
}

func (vP *VideoProcessor) SwitchState(id int32) {
	vP.Switcher <- id
}

func GetVideoProcessor(numOfCores int) *VideoProcessor {
	return &VideoProcessor{
		CPUs:     numOfCores,
		Chanel:   make(chan struct{}, numOfCores),
		XMLfile:  "../haarcascade_frontalface_default.xml",
		Videos:   make(map[int32]Video),
		Switcher: make(chan int32)}
}

func (vP *VideoProcessor) updateVideo(id int32, name string, status VideoStatus, percentage float64) {
	mutex := sync.RWMutex{}
	mutex.Lock()
	copy, _ := vP.Videos[id]
	copy.Name = name
	copy.Percentage = percentage
	copy.Status = status
	vP.Videos[id] = copy
	mutex.Unlock()
}

// fileName is used nothing, but logging file name
func (vP *VideoProcessor) ProcessVideo(ctx context.Context, videoFile, xmlFile, fileName string, wg *sync.WaitGroup) {
	var id = vP.processId.Add(1)
	status := true
	vP.updateVideo(id, fileName, 0, 0.0)
	vP.Chanel <- struct{}{}
	vP.updateVideo(id, fileName, 1, 0.0)
	//open video file
	video, err := gocv.VideoCaptureFile(videoFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer video.Close()

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
		vP.updateVideo(id, fileName, 2, 0.0)
		return
	}

	fmt.Printf("start reading video from: %s\n", videoFile)

	for {
		var progress = float64(frame_counter) / total_frames * 100
		select {
		case sw := <-vP.Switcher:
			if sw == id {
				status = !status
			}
			if status {
				log.Printf("goroutine: %d, processId: %d - %.2f%%: Processing of %s was resumed\n", gr, id, progress, fileName)
				vP.updateVideo(id, fileName, 1, progress)
			} else {
				log.Printf("goroutine: %d, processId: %d - %.2f%%: Processing of %s was paused\n", gr, id, progress, fileName)
				vP.updateVideo(id, fileName, 5, progress)
			}
		//if request is canceled, goroutine is shutting down
		case <-ctx.Done():
			vP.updateVideo(id, fileName, 3, progress)
			wg.Done()
			return
		default:
			if status {
				vP.updateVideo(id, fileName, 1, progress)
				if ok := video.Read(&img); !ok {
					fmt.Printf("cannot read video from file %s\n", videoFile)
					vP.updateVideo(id, fileName, 4, progress)
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
