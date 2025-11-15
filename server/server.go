package server

// Importing packages
import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"server/api"
	"server/model"
	"strconv"
	"strings"
	"sync"
	"time"
)

var server *http.Server
var serverMutex sync.Mutex

func InitProblems(path string) error {
	return parse_problems(path)
}

func Init(port uint16) error {
	serverMutex.Lock()
	defer serverMutex.Unlock()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/challenge", api.RouteGET_CurrentChallenge)
	mux.HandleFunc("/api/check_solution", api.RoutePOST_CheckSolution)
	mux.HandleFunc("/api/submit", api.RoutePOST_Submit)
	mux.HandleFunc("/api/join", api.RoutePOST_JoinUser)
	mux.HandleFunc("/api/get_users", api.RoutePOST_GetUsers)
	mux.HandleFunc("/api/get_submissions", api.RoutePOST_GetSubmissions)
	mux.HandleFunc("/api/get_code_reviews", api.RouteGET_GetCodeReviews)
	mux.HandleFunc("/api/speed_leaderboard", api.RouteGET_SpeedLeaderboard)
	mux.HandleFunc("/api/quality_leaderboard", api.RoutePOST_GetSubmissions)
	mux.HandleFunc("/api/get_state", api.RouteGET_GetState)
	mux.HandleFunc("/api/add_code_review", api.RoutePOST_AddCodeReview)

	server = &http.Server{
		Addr:    ":" + strconv.Itoa(int(port)),
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		}
	}()

	return nil
}

func Terminate() {
	serverMutex.Lock()
	defer serverMutex.Unlock()

	if server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			fmt.Println("Server Shutdown Error:", err)
		}
	}
}

func json_to_problem(value interface{}, itr uint32) (*model.Problem, error) {
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

	diffStr, ok := problemJSON["Difficulty"].(string)
	if !ok || diffStr == "" {
		return nil, fmt.Errorf("invalid problem: missing or invalid 'Difficulty'")
	}

	var diff model.ProblemDifficulty
	if diffStr == "Easy" {
		diff = model.Easy
	} else if diffStr == "Medium" {
		diff = model.Medium
	} else if diffStr == "Hard" {
		diff = model.Hard
	} else {
		//return fmt.Errorf("invalid problem: missing or invalid 'Difficulty': ", diffStr, ". Valid options are: \"Easy\", \"Medium\", \"Hard\""), nil
	}

	testCasesMap := problemJSON["TestCases"].([]interface{})

	var testCases []model.TestCase
	for _, val := range testCasesMap {
		var testCase model.TestCase
		testCaseJSON, ok := val.(map[string]interface{})
		// get case sensitive
		caseSensitive, ok := testCaseJSON["CaseSensitive"].(bool)
		if !ok || diffStr == "" {
			return nil, fmt.Errorf("invalid problem: missing or invalid field 'CaseSensitive'")
		}
		testCase.CaseSensitive = caseSensitive

		// get input
		inputJSONmap, ok := testCaseJSON["Input"].(map[string]interface{})

		// Marshal the map into a JSON byte slice
		jsonBytes, err := json.Marshal(inputJSONmap)
		if err != nil {
			return nil, fmt.Errorf("invalid input JSON structure")
		}

		testCase.Input = string(jsonBytes)

		if !ok || testCase.Input == "" {
			return nil, fmt.Errorf("invalid problem: missing or invalid field 'Input'")
		}

		// get output
		outputStr, ok := testCaseJSON["Output"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid problem: missing or invalid field 'Output'")
		}
		testCase.OutputJSON = outputStr

		testCases = append(testCases, testCase)
	}

	problem := &model.Problem{
		Header: model.ProblemHeader{
			Name:        nameStr,
			Description: headerStr,
		},
		Difficulty: diff,
		Id:         uint16(itr),
		TestCases:  testCases,
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

	bytes, _ := io.ReadAll(jsonFile)

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
		var problem *model.Problem
		problem, err = json_to_problem(value, itr)

		if err != nil {
			return err
		}
		model.ProblemList = append(model.ProblemList, *problem)

		itr++
	}
	return nil
}

func parse_problems(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	fmt.Println("Loaded problems:")
	for _, e := range entries {
		if !e.IsDir() {
			parse_problem_file(dir + "/" + e.Name())
		}
		fmt.Println("\t", strings.Split(e.Name(), ".")[0])
	}

	return nil
}
