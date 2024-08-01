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

// ShowAccount godoc
// @Summary      Show an account
// @Description  get string by ID
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Account ID"
// @Success      200  {object}  model.Account
// @Failure      400  {object}  httputil.HTTPError
// @Failure      404  {object}  httputil.HTTPError
// @Failure      500  {object}  httputil.HTTPError
// @Router       /accounts/{id} [get]
// router.GET("/upload", UploadHtml)
func (service *VideoService) UploadHtml(c *gin.Context) {
	c.HTML(200, "main.html", gin.H{})
}

// ListAccounts godoc
// @Summary      List accounts
// @Description  get accounts
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Param        q    query     string  false  "name search by q"  Format(email)
// @Success      200  {array}   model.Account
// @Failure      400  {object}  httputil.HTTPError
// @Failure      404  {object}  httputil.HTTPError
// @Failure      500  {object}  httputil.HTTPError
// @Router       /accounts [get]
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
	go service.vP.ProcessVideo(c.Request.Context(), path, service.vP.XMLfile, file.Filename, &wg)
	wg.Wait()
	if c.Request.Context().Err() == context.Canceled {
		log.Printf("Request for '%s' was aborted", file.Filename)
		return
	}
	c.String(http.StatusOK, fmt.Sprintf("'%s' uploaded!", file.Filename))
}
