package main

import (
	_ "go_cv_test/docs"
	"go_cv_test/internal/handlers"
)

func main() {
	s := handlers.GetService()
	s.Run()
}
