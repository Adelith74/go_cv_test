package handlers

import (
	"log"
	"runtime"

	model "go_cv_test/internal/model"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type VideoService struct {
	vP *model.VideoProcessor
}

func GetService() VideoService {
	return VideoService{vP: model.GetVideoProcessor(runtime.NumCPU())}
}

//	@title			Swagger Example API
//	@version		1.0
//	@description	This is a sample server celler server.
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

// @host		localhost:8080
// @BasePath	/api/v1
func (service *VideoService) Run() {
	router := gin.Default()
	router.Static("/static", "../web/static")
	router.LoadHTMLFiles("../web/static/templates/main.html")
	log.Printf("Videos will be proceed with %d cores", service.vP.CPUs)
	// Set a lower memory limit for multipart forms (default is 32 MiB)
	router.MaxMultipartMemory = 8 << 20 // 8 MiB

	v1 := router.Group("/api/v1")
	{
		upload := v1.Group("/upload")
		{
			upload.POST("", service.UploadVideo)
			upload.GET("", service.UploadHtml)
		}
		state := v1.Group("/switch_state")
		{
			state.POST("", service.SwitchState)
		}
		status := v1.Group("/status")
		{
			status.GET("", service.GetStatus)
		}
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.Run(":8080")
}
