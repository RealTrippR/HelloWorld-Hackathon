package server

// Importing packages
import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"sync"
)

type ProblemDifficulty int

const (
	Easy = iota
	Medium
	Hard
)

type ProblemHeader struct {
	name        string
	description string
}

type Problem struct {
	header         ProblemHeader
	difficulty     ProblemDifficulty
	id             uint16
	objective      string
	expectedOutput map[string]interface{}
}

type clientInfo struct {
	useTLS bool
	conn   net.Conn
}

var crtPath = "server.crt"
var keyPath = "server.key"

var ln net.Listener
var cer tls.Certificate
var config tls.Config

var serverMutex sync.Mutex

var problemList []Problem

func InitProblems(path string) error {
	return parse_problem_file(path)
}

func Init() error {
	serverMutex.Lock()
	defer serverMutex.Unlock()

	var err error
	// https://www.baeldung.com/linux/crt-key-files
	cer, err = tls.LoadX509KeyPair(crtPath, keyPath)

	if err != nil {
		log.Println(err, "Failed to load X509 key pair, crtPath: ", crtPath, ", keyPath: ", keyPath)
		return err
	}

	config = tls.Config{Certificates: []tls.Certificate{cer}}

	ln, err = tls.Listen("tcp", ":443", &config)
	if err != nil {
		log.Println(err)
		return err
	}

	go serverThread()

	return nil
}

func serverThread() {
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		clientInfo := new(clientInfo)
		clientInfo.conn = conn
		clientInfo.useTLS = true
		go handleConnection(clientInfo)
	}
}

func Terminate() {
	serverMutex.Lock()
	ln.Close()
	serverMutex.Unlock()
}

func forwardRequest(client *clientInfo, port uint16) {

}

func handleConnection(client *clientInfo) {
	defer client.conn.Close()
	r := bufio.NewReader(client.conn)
	for {
		msg, err := r.ReadString('\n')
		if err != nil {
			log.Println(err)
			if (errors.Is(err, tls.RecordHeaderError{})) {

			} else {
				return
			}
		}

		println(msg)

		n, err := client.conn.Write([]byte("world\n"))
		if err != nil {
			log.Println(n, err)
			return
		}
	}
}

func json_to_problem(value interface{}, itr uint32) (*Problem, error) {
	problemJSON, ok := value.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid problem in problems list; problem is not an object.")
	}

	nameStr, ok := problemJSON["Name"].(string)
	if !ok || nameStr == "" {
		return nil, fmt.Errorf("invalid problem: missing or invalid 'Name'")
	}

	headerStr, ok := problemJSON["Header"].(string)
	if !ok || headerStr == "" {
		return nil, fmt.Errorf("invalid problem: missing or invalid 'Header'")
	}

	expectedOutput, ok := problemJSON["ExpectedOutput"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid problem: missing or invalid 'ExpectedOutput'")
	}

	diffStr, ok := problemJSON["Difficulty"].(string)
	if !ok || diffStr == "" {
		return nil, fmt.Errorf("invalid problem: missing or invalid 'Difficulty'")
	}

	var diff ProblemDifficulty
	if diffStr == "Easy" {
		diff = Easy
	} else if diffStr == "Medium" {
		diff = Medium
	} else if diffStr == "Hard" {
		diff = Hard
	} else {
		return nil, fmt.Errorf("invalid problem: missing or invalid 'Difficulty': Valid options are: \"Easy\", \"Medium\", \"Hard\"")
	}

	objective, ok := problemJSON["Objective"].(string)
	if !ok || diffStr == "" {
		return nil, fmt.Errorf("invalid problem: missing or invalid 'Objective'")
	}

	problem := &Problem{
		header: ProblemHeader{
			name:        nameStr,
			description: headerStr,
		},
		difficulty:     diff,
		id:             uint16(itr),
		objective:      objective,
		expectedOutput: expectedOutput,
	}

	return problem, nil
}

func parse_problem_file(path string) error {
	jsonFile, err := os.Open(path)
	if err != nil {
		fmt.Println("Failed to open JSON problems file! err: ", err)
		return err
	}
	defer jsonFile.Close()

	bytes, _ := ioutil.ReadAll(jsonFile)

	var result map[string]interface{}
	err = json.Unmarshal([]byte(bytes), &result)
	if err != nil {
		return fmt.Errorf("invalid problems file:  bad JSON structure")
	}

	_, ok := result["Problems"]
	if !ok {
		return fmt.Errorf("invalid problems file:  bad JSON structure")
	}
	problemsMap := result["Problems"].([]interface{})

	var itr uint32 = 0
	for _, value := range problemsMap {
		var problem *Problem
		problem, err = json_to_problem(value, itr)

		if err != nil {
			return err
		}
		problemList = append(problemList, *problem)

		problem_to_json(problem)

		itr++
	}
	return nil
}

func problem_to_json(problem *Problem) string {
	jsonBytes, err := json.Marshal(*problem)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return ""
	}

	jsonString := string(jsonBytes)
	fmt.Println(jsonString)
	return string(jsonBytes)
}
