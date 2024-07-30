package routers

import (
	"context"
	"fmt"
	videoProcessor "go_cv_test/internal/model"
	"log"
	"net/http"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/gin-gonic/gin"
)

type VideoService struct {
	vP *videoProcessor.VideoProcessor
}

func GetService() VideoService {
	return VideoService{vP: videoProcessor.GetVideoProcessor(runtime.NumCPU())}
}

func (service *VideoService) Run() {
	router := gin.Default()
	router.Static("/static", "../web/static")
	router.LoadHTMLFiles("../web/static/templates/main.html")
	log.Printf("Videos will be proceed with %d cores", service.vP.CPUs)
	// Set a lower memory limit for multipart forms (default is 32 MiB)
	router.MaxMultipartMemory = 8 << 20 // 8 MiB

	router.POST("/upload", func(c *gin.Context) {
		// single file
		file, _ := c.FormFile("file")
		log.Println(file.Filename + " was recieved")
		// Upload the file to specific dst.
		filename := filepath.Base(file.Filename)
		path := "../files/" + filename
		if err := c.SaveUploadedFile(file, path); err != nil {
			c.String(http.StatusBadRequest, "upload file err: %s", err.Error())
			return
		}

		log.Printf("Start processing file...")
		wg := sync.WaitGroup{}
		wg.Add(1)
		go service.vP.ProcessVideo(c.Request.Context(), path, service.vP.XMLfile, file.Filename, &wg)
		wg.Wait()
		if c.Request.Context().Err() == context.Canceled {
			log.Printf("Request for '%s' was aborted", file.Filename)
			return
		}
		c.String(http.StatusOK, fmt.Sprintf("'%s' uploaded!", file.Filename))
	})

	router.GET("/upload", func(c *gin.Context) {
		c.HTML(200, "main.html", gin.H{})
	})
	router.Run(":8080")
}
