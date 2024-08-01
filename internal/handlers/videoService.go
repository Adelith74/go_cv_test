package handlers

import (
	"log"
	"net/http"
	"runtime"
	"strconv"

	"go_cv_test/docs"
	model "go_cv_test/internal/model"

	"github.com/gin-gonic/gin"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type VideoService struct {
	vP *model.VideoProcessor
}

func GetService() VideoService {
	return VideoService{vP: model.GetVideoProcessor(runtime.NumCPU())}
}

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html
func (service *VideoService) Run() {
	router := gin.Default()
	router.Static("/static", "../web/static")
	router.LoadHTMLFiles("../web/static/templates/main.html")
	log.Printf("Videos will be proceed with %d cores", service.vP.CPUs)
	// Set a lower memory limit for multipart forms (default is 32 MiB)
	router.MaxMultipartMemory = 8 << 20 // 8 MiB

	docs.SwaggerInfo.Title = "Swagger Example API"
	docs.SwaggerInfo.Description = "This is a sample server Petstore server."
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = "petstore.swagger.io"
	docs.SwaggerInfo.BasePath = "/v2"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}

	//this route is used for pausing and unpausing videos from proceeding, paused goroutines wont be deleted
	router.POST("/switch_state", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Query("id"))
		if err != nil {
			c.String(http.StatusBadRequest, "Unable to process id")
		}
		service.vP.SwitchState(int32(id))
		if err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
		} else {
			c.JSON(http.StatusOK, gin.H{})
		}
	})

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	router.Run(":8080")
}
