package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"server/model"
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

	caseIdx := received["TestCase"].(float64)

	problem := model.GetCurrentProblem()
	// Example: expected JSON object
	if int(caseIdx) >= len(problem.TestCases) {
		w.WriteHeader(http.StatusBadRequest)
	}

	expected := model.GetCurrentProblem().TestCases[int(caseIdx)].OutputJSON

	jsonBytesExpected, err := json.Marshal(expected)
	fmt.Println("EXPECTED: ", string(jsonBytesExpected))

	jsonBytesRecevied, err := json.Marshal(received["Output"])
	fmt.Println("RECEIVED: ", string(jsonBytesRecevied))

	// Check equality
	if string(jsonBytesRecevied) == string(jsonBytesExpected) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"correct": true}`))
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"correct": false}`))
	}
}

func RoutePOST_joinUser(w http.ResponseWriter, r *http.Request) {
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

}
