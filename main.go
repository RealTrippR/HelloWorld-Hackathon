package main

import (
	"bufio"
	"fmt"
	"os"
	"server/server"
)

func main() {
	err := server.InitProblems("problems/problems.json")
	if err != nil {
		fmt.Println("Error parsing JSON problems file.")
		return
	}

	err = server.Init()
	if err != nil {
		fmt.Println("Error, closing server.")
		return
	}

	fmt.Print("Press any key to quit the server.")
	reader := bufio.NewReader(os.Stdin)
	_, _ = reader.ReadString('\n')
	server.Terminate()
}
