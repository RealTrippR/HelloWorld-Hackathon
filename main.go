package main

import (
	"bufio"
	"fmt"
	"os"
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

	fmt.Println("Server open on port", port, ".\nPress any key to quit the server.")
	reader := bufio.NewReader(os.Stdin)
	_, _ = reader.ReadString('\n')
	server.Terminate()
}
