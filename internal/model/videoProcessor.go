package videoProcessor

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"strconv"
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

func (vP *VideoProcessor) WebDetect() {
	if len(os.Args) < 3 {
		fmt.Println("How to run:\n\tfacedetect [camera ID] [classifier XML file]")
		return
	}

	// parse args
	deviceID, _ := strconv.Atoi(os.Args[1])
	xmlFile := os.Args[2]

	// open webcam
	webcam, err := gocv.VideoCaptureDevice(int(deviceID))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer webcam.Close()

	// open display window
	window := gocv.NewWindow("Face Detect")
	defer window.Close()

	// prepare image matrix
	img := gocv.NewMat()
	defer img.Close()

	// color for the rect when faces detected
	blue := color.RGBA{0, 0, 255, 0}

	// load classifier to recognize faces
	classifier := gocv.NewCascadeClassifier()
	defer classifier.Close()

	if !classifier.Load(xmlFile) {
		fmt.Printf("Error reading cascade file: %v\n", xmlFile)
		return
	}

	fmt.Printf("start reading camera device: %v\n", deviceID)
	for {
		if ok := webcam.Read(&img); !ok {
			fmt.Printf("cannot read device %d\n", deviceID)
			return
		}
		if img.Empty() {
			continue
		}

		// detect faces
		rects := classifier.DetectMultiScale(img)
		fmt.Printf("found %d faces\n", len(rects))

		// draw a rectangle around each face on the original image,
		// along with text identifying as "Human"
		for _, r := range rects {
			gocv.Rectangle(&img, r, blue, 3)

			size := gocv.GetTextSize("Human", gocv.FontHersheyPlain, 1.2, 2)
			pt := image.Pt(r.Min.X+(r.Min.X/2)-(size.X/2), r.Min.Y-2)
			gocv.PutText(&img, "Human", pt, gocv.FontHersheyPlain, 1.2, blue, 2)
		}

		// show the image in the window, and wait 1 millisecond
		window.IMShow(img)
		if window.WaitKey(1) >= 0 {
			break
		}
	}
}

// fileName is used nothing, but logging file name
func (vP *VideoProcessor) ProcessVideo(videoFile, xmlFile, fileName string) {
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
		if ok := video.Read(&img); !ok {
			fmt.Printf("cannot read video from file %s\n", videoFile)
			return
		}
		if img.Empty() {
			continue
		}

		// detect faces
		rects := classifier.DetectMultiScale(img)
		fmt.Printf("%d: found %d faces on frame %d of %s\n", gr, len(rects), frame_counter, fileName)
		frame_counter++
	}

}
