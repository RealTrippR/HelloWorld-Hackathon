package api

import (
	"encoding/json"
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

	model.Mutex.Lock()
	defer model.Mutex.Unlock()

	authed, _ := model.IsAuthedRequest(received)
	if !authed {
		http.Error(w, "Invalid UserId", http.StatusBadRequest)
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
	//fmt.Println("EXPECTED: ", string(jsonBytesExpected))

	jsonBytesRecevied, err := json.Marshal(received["Output"])
	//fmt.Println("RECEIVED: ", string(jsonBytesRecevied))

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

func RoutePOST_Submit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed: Expected POST", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Decode the JSON body into a map
	var received map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&received)
	if err != nil {
		http.Error(w, "{\"Error:\":\"Invalid JSON payload\"}", http.StatusBadRequest)
		return
	}

	authed, userId := model.IsAuthedRequest(received)
	if !authed {
		http.Error(w, "{\"Error:\":\"Invalid JSON payload\"}", http.StatusBadRequest)
		return
	}

	sourceFileMap, ok := received["SourceFiles"].([]interface{})
	if !ok {
		http.Error(w, "Missing or invalid field 'SourceFiles'", http.StatusBadRequest)
		return
	}

	var srcFileList []model.SourceFile
	for _, value := range sourceFileMap {
		var srcFile model.SourceFile
		// parse source files
		sourceFileJSON, ok := value.(map[string]interface{})
		if !ok {
			http.Error(w, "{\"Error:\":\"Improperly structured list of sourcefiles. Expects an array of objects: [{'Name': string, 'Code': string}]}\"", http.StatusBadRequest)
			return
		}

		nameStr, ok := sourceFileJSON["Name"].(string)
		if !ok || nameStr == "" {
			http.Error(w, "{\"Error:\":\"Improperly structured list of sourcefiles. Expects an array of objects: [{'Name': string, 'Code': string}]}\"", http.StatusBadRequest)
			return
		}

		codeStr, ok := sourceFileJSON["Code"].(string)
		if !ok || codeStr == "" {
			http.Error(w, "{\"Error:\":\"Improperly structured list of sourcefiles. Expects an array of objects: [{'Name': string, 'Code': string}]}\"", http.StatusBadRequest)
			return
		}

		srcFile.Name = nameStr
		srcFile.Code = codeStr
		srcFileList = append(srcFileList, srcFile)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{\"Error\":\"Success\"}"))
	model.AddSubmission(userId, srcFileList)
}

// RoutePOST_GetUsers returns a list of all registered users
func RoutePOST_GetUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	model.Mutex.Lock()
	defer model.Mutex.Unlock()

	// Build a response array
	users := make([]interface{}, 0, len(model.Users))
	for _, u := range model.Users {
		users = append(users, map[string]interface{}{
			"Name": u.Name,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// RoutePOST_GetSubmissions returns the submissions for a given UserID
func RoutePOST_GetSubmissions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed: Expected GET", http.StatusMethodNotAllowed)
		return
	}

	// Decode request body
	var received map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	auth, uId := model.IsAuthedRequest(received)
	if !auth {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	model.Mutex.Lock()
	defer model.Mutex.Unlock()

	sourceFiles, ok := model.Submissions[uId]
	if !ok {
		sourceFiles = []model.SourceFile{} // return empty slice if none
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sourceFiles)
}

func RoutePOST_JoinUser(w http.ResponseWriter, r *http.Request) {
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

	model.Mutex.Lock()
	defer model.Mutex.Unlock()
	if model.IsUsernameTaken(username) {
		w.Write([]byte(`{"Error":"name taken"}`))
		return
	}

	var id int64

	err, id = model.AddUser(username)

	if err != nil {
		w.Write([]byte(`{"Error":"err"}`))
	}

	w.Header().Set("Content-Type", "application/json")
	resp := map[string]interface{}{
		"Error":  "success",
		"UserID": id, // uint64, stays a number
	}

	json.NewEncoder(w).Encode(resp)
}

func RouteGET_SpeedLeaderboard(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed: Expected GET", http.StatusMethodNotAllowed)
		return
	}

}

func RouteGET_QualityLeaderboard(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed: Expected GET", http.StatusMethodNotAllowed)
		return
	}

}

func RouteGET_GetState(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed: Expected GET", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if model.GetCycleState() == model.Coding {
		w.Write([]byte("{\"State\":\"coding\"}"))
		return
	} else {
		w.Write([]byte("{\"State\":\"reviewing\"}"))
		return
	}
}
