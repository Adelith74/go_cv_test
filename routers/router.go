package routers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type VideoService struct {
}

func GetService() VideoService {
	return VideoService{}
}

func (service *VideoService) Run() {
	router := gin.Default()
	// Set a lower memory limit for multipart forms (default is 32 MiB)
	router.MaxMultipartMemory = 8 << 20 // 8 MiB
	router.POST("/upload", func(c *gin.Context) {
		// single file
		file, _ := c.FormFile("file")
		log.Println(file.Filename)

		// Upload the file to specific dst.
		c.SaveUploadedFile(file, "../files/")

		c.String(http.StatusOK, fmt.Sprintf("'%s' uploaded!", file.Filename))
	})
	router.Run(":8080")
}
