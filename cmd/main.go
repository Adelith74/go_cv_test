package main

import "go_cv_test/routers"

func main() {
	//with command below we are able to set and get max amount of PROCS
	//max := runtime.GOMAXPROCS(0)
	//with command below we are able to find amount of CPUs on local machine
	//runtime.NumCPU()
	s := routers.GetService()
	s.Run()
}
