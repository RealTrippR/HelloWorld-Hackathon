package model

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"sync"
	"time"
)

type ProblemDifficulty int

const (
	Easy = iota
	Medium
	Hard
)

type ProblemHeader struct {
	Name        string
	Description string
}

type Problem struct {
	Header     ProblemHeader
	Difficulty ProblemDifficulty
	Id         uint16
	Objective  string
	TestCases  []TestCase
}

type SourceFile struct {
	Name string
	Code string
}
type TestCase struct {
	Input         string
	OutputJSON    map[string]interface{}
	CaseSensitive bool
}

type User struct {
	Name      string
	privateId int64
}

type CycleTime int

const (
	Coding = iota
	Review
)

type CycleState struct {
	LastCycleTime     time.Time
	currentProblemIdx uint32
	cycle             CycleTime
	codingDurMins     float64 // time of the coding cycle, in minutes
	reviewDurMins     float64 // time of the review cycle in minutes
}

var Users map[int64]User // LOOKUP BY PRIVATE ID

var Submissions map[int64][]SourceFile // LOOKUP BY PRIVATE ID

var ProblemList []Problem

var cycleState CycleState

var Mutex sync.Mutex

func Init() {
	Users = make(map[int64]User)
	Submissions = make(map[int64][]SourceFile)
	cycleState.currentProblemIdx = 0
	cycleState.LastCycleTime = time.Now()
	cycleState.codingDurMins = 30.0
	cycleState.reviewDurMins = 10.0
}

func Tick() {

	elapsed := time.Since(cycleState.LastCycleTime)
	minutes := elapsed.Minutes()

	if cycleState.cycle == Coding && minutes > cycleState.codingDurMins {
		cycleState.LastCycleTime = time.Now()
		cycleState.cycle = Review
	} else if cycleState.cycle == Review && minutes > cycleState.reviewDurMins {
		// PROCEED TO NEXT PROBLEM
		cycleState.cycle = Coding
		CycleProblem()
	}
}

func CycleProblem() {
	Submissions = make(map[int64][]SourceFile)
	cycleState.LastCycleTime = time.Now()
	cycleState.currentProblemIdx++
	if cycleState.currentProblemIdx >= uint32(len(ProblemList)) {
		cycleState.currentProblemIdx = 0
	}
}

func GetCurrentProblem() *Problem {
	return &ProblemList[cycleState.currentProblemIdx]
}

func IsUsernameTaken(name string) bool {
	for _, val := range Users {
		if val.Name == name {
			return true
		}
	}
	return false
}

func generateSecureRandomInt64() (int64, error) {
	var b [8]byte
	_, err := rand.Read(b[:])
	if err != nil {
		return 0, fmt.Errorf("failed to read from crypto/rand: %w", err)
	}

	return int64(binary.LittleEndian.Uint64(b[:])), nil
}

func IsAuthedRequest(received map[string]interface{}) (bool, int64) {
	return true, 0 // DEBUG ONLY

	// Extract UserID
	raw, ok := received["UserID"]
	if !ok {
		return false, 0
	}
	uId := int64(raw.(float64))

	if IsValidUserId(uId) {
		return true, uId
	} else {
		return false, 0
	}
}

func IsValidUserId(userId int64) bool {
	_, ok := Users[userId]
	return ok
}

func AddSubmission(uId int64, sourceFiles []SourceFile) {
	Submissions[uId] = sourceFiles
}

func AddUser(name string) (error, int64) {
	var u User
	var err error
	u.Name = name
	u.privateId, err = generateSecureRandomInt64()
	if err != nil {
		return err, 0
	}
	_, ok := Users[u.privateId]
	if ok {
		// key exists
		for !ok {
			u.privateId, err = generateSecureRandomInt64()
			if err != nil {
				return err, 0
			}
			_, ok = Users[u.privateId]
		}
	}

	Users[u.privateId] = u
	return nil, u.privateId
}
