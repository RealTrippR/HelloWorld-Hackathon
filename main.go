package main

import (
	"fmt"
	"server/model"
	"server/server"
)

var port uint16 = 3000

func main() {
	model.Init()

	err := server.InitProblems("problems")
	if err != nil {
		fmt.Println("Error parsing JSON problems file.")
		return
	}

	err = server.Init(port)
	if err != nil {
		fmt.Println("Error, closing server.")
		return
	}

	fmt.Println("Server open on port", port, ".\nPress any ctrl+C to quit the server.")

	for {
		model.Mutex.Lock()
		model.Tick()
		model.Mutex.Unlock()
	}

	server.Terminate()
}
