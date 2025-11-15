package model

import "time"

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

type TestCase struct {
	Input         string
	OutputJSON    map[string]interface{}
	CaseSensitive bool
}

type User struct {
	nickname  string
	privateId uint64
}

var Users map[uint64]User // LOOKUP BY PRIVATE ID

var ProblemList []Problem
var currentProblemIdx uint32

var LastCycleTime time.Time

func CycleProblem() {
	LastCycleTime = time.Now()
	currentProblemIdx++
	if currentProblemIdx >= uint32(len(ProblemList)) {
		currentProblemIdx = 0
	}
}

func GetCurrentProblem() *Problem {
	return &ProblemList[currentProblemIdx]
}
