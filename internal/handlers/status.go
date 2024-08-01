package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetStatus godoc
//
//	@Summary		Get status of a video
//	@Description	Return current status, 0 - queue, 1 - processing, 2 - error, 3 - canceled, 4 - successful, 5 - paused
//	@Accept			json
//	@Produce		json
//	@Param			id	query		int	true	"id"
//	@Success		200	{object}	model.Video
//	@Failure		400	{object}	string
//	@Router			/get_status [post]
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
