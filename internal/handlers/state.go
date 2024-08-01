package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// SwitchState godoc
//
//	@Summary		Switch state of a video
//	@Description	Switch by video ID
//	@Accept			json
//	@Produce		json
//	@Param			id	query		int	true	"id"
//	@Success		200	{object}	int
//	@Failure		400	{object}	int
//	@Router			/switch_state [post]
//
// this route is used for pausing and unpausing videos from proceeding, paused goroutines wont be deleted
func (service *VideoService) SwitchState(c *gin.Context) {
	id, err := strconv.Atoi(c.Request.URL.Query().Get("id"))
	if err != nil {
		c.String(http.StatusBadRequest, "Unable to process id")
		return
	}
	err = service.vP.SwitchState(int32(id))
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	} else {
		c.JSON(http.StatusOK, gin.H{})
		return
	}
}
