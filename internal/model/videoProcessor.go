package videoProcessor

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"

	"gocv.io/x/gocv"
)

type VideoProcessor struct {
	CPUs      int
	Chanel    chan struct{}
	XMLfile   string
	grCounter atomic.Int32
}

func GetVideoProcessor(numOfCores int) *VideoProcessor {
	return &VideoProcessor{CPUs: numOfCores, Chanel: make(chan struct{}, numOfCores), XMLfile: "../haarcascade_frontalface_default.xml"}
}

// fileName is used nothing, but logging file name
func (vP *VideoProcessor) ProcessVideo(ctx context.Context, videoFile, xmlFile, fileName string, wg *sync.WaitGroup) {
	vP.Chanel <- struct{}{}
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
		return
	}

	fmt.Printf("start reading video from: %s\n", videoFile)

	for {
		var progress = float64(frame_counter) / total_frames * 100
		select {
		case <-ctx.Done():
			wg.Done()
			return
		default:
			if ok := video.Read(&img); !ok {
				fmt.Printf("cannot read video from file %s\n", videoFile)
				wg.Done()
				return
			}
			if img.Empty() {
				continue
			}

			// detect faces
			rects := classifier.DetectMultiScale(img)
			log.Printf("%d - %.2f%%: found %d faces on frame %d of %s\n", gr, progress, len(rects), frame_counter, fileName)
			frame_counter++
		}
	}

}
