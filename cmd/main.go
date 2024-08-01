package main

import (
	"go_cv_test/routers"
)

// @BasePath /api/v1

// PingExample godoc
// @Summary ping example
// @Schemes
// @Description do ping
// @Tags example
// @Accept json
// @Produce json
// @Success 200 {string} Helloworld
// @Router /example/helloworld [get]

func main() {
	s := routers.GetService()
	s.Run()
}
