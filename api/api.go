package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"server/model"
	"strings"
)

func RouteGET_CurrentChallenge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed: Expected GET", http.StatusMethodNotAllowed)
		return
	}

	// Convert struct to JSON
	jsonData, err := json.Marshal(model.GetCurrentProblem())
	if err != nil {
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Write JSON to response
	w.Write(jsonData)
}

func RoutePOST_CheckSolution(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed: Expected POST", http.StatusMethodNotAllowed)
		return
	}

	// Decode the JSON body into a map
	var received map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&received)
	if err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if !model.IsAuthedRequest(received) {
		http.Error(w, "Invalid UserId", http.StatusBadRequest)
		return
	}

	// Extract UserID
	raw, ok := received["UserID"]
	if !ok {
		http.Error(w, "Missing UserID", http.StatusBadRequest)
		return
	}
	uId := int64(raw.(float64))
	if !model.IsValidUserId(uId) {
		http.Error(w, "Invalid UserID", http.StatusBadRequest)
		return
	}

	caseIdx := received["TestCase"].(float64)

	problem := model.GetCurrentProblem()
	// Example: expected JSON object
	if int(caseIdx) >= len(problem.TestCases) {
		w.WriteHeader(http.StatusBadRequest)
	}

	testCase := model.GetCurrentProblem().TestCases[int(caseIdx)]
	expected := testCase.OutputJSON

	jsonBytesExpected, err := json.Marshal(expected)
	fmt.Println("EXPECTED: ", string(jsonBytesExpected))

	jsonBytesRecevied, err := json.Marshal(received["Output"])
	fmt.Println("RECEIVED: ", string(jsonBytesRecevied))

	var isCorrect bool
	if testCase.CaseSensitive {
		isCorrect = string(jsonBytesRecevied) == string(jsonBytesExpected)
	} else {
		isCorrect = strings.EqualFold(string(jsonBytesRecevied), string(jsonBytesExpected))
	}
	// Check equality
	if isCorrect {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"Correct": true}`))
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"Correct": false}`))
	}
}

func RoutePOST_joinUser(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed: Expected POST", http.StatusMethodNotAllowed)
		return
	}

	// Decode the JSON body into a map
	var received map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&received)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	username, ok := received["Username"].(string)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if model.IsUsernameTaken(username) {
		w.Write([]byte(`{"Error":"name taken"}`))
	}

	var id int64
	err, id = model.AddUser(username)
	if err != nil {
		w.Write([]byte(`{"Error":"err"}`))
	}

	resp := map[string]interface{}{
		"Error":  "success",
		"UserID": id, // uint64, stays a number
	}

	json.NewEncoder(w).Encode(resp)
}
