package main

import service "go_cv_test/routers"

func main() {
	s := service.GetService()
	s.Run()
}
