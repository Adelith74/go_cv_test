package main

import (
	_ "go_cv_test/docs"
	"go_cv_test/internal/handlers"
)

// @BasePath /api/v1

func main() {
	s := handlers.GetService()
	s.Run()
}
