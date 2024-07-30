package main

import "go_cv_test/routers"

func main() {
	s := routers.GetService()
	s.Run()
}
