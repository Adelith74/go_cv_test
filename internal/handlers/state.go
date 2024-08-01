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
//	@Param			id	path		int	true	"id"	Format(int32)
//	@Success		200	{object}	int
//	@Failure		400	{object}	int
//	@Router			/switch_state [post]
//
// this route is used for pausing and unpausing videos from proceeding, paused goroutines wont be deleted
func (service *VideoService) SwitchState(c *gin.Context) {
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
}
