package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (service *VideoService) GetStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Query("id"))
	if err != nil {
		c.String(http.StatusBadRequest, "Unable to process id")
	}
	video, err := service.vP.GetVideo(int32(id))
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
	} else {
		c.JSON(http.StatusOK, video)
	}
}
