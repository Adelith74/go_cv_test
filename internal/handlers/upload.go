package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/gin-gonic/gin"
)

// router.GET("/upload", UploadHtml)
func (service *VideoService) UploadHtml(c *gin.Context) {
	c.HTML(200, "main.html", gin.H{})
}

// //router.POST("/upload", func(c *gin.Context) {
func (service *VideoService) UploadVideo(c *gin.Context) {
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
	ctx, cancel := context.WithCancelCause(c.Request.Context())
	go service.vP.ProcessVideo(ctx, cancel, path, service.vP.XMLfile, file.Filename, &wg)
	wg.Wait()
	if ctx.Err() == context.Canceled {
		c.String(http.StatusBadRequest, fmt.Sprintf("Request for '%s' was aborted due to: %s", file.Filename, ctx.Err().Error()))
		return
	}
	c.String(http.StatusOK, fmt.Sprintf("'%s' was processed successfully!", file.Filename))
}
