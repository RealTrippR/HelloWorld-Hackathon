package model

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
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

type TestCase struct {
	Input         string
	OutputJSON    map[string]interface{}
	CaseSensitive bool
}

type User struct {
	nickname  string
	privateId int64
}

var Users map[int64]User // LOOKUP BY PRIVATE ID

var ProblemList []Problem
var currentProblemIdx uint32

var LastCycleTime time.Time

func Init() {
	Users = make(map[int64]User)
}

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

func IsUsernameTaken(name string) bool {
	for _, val := range Users {
		if val.nickname == name {
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

func IsAuthedRequest(received map[string]interface{}) bool {
	// Extract UserID
	raw, ok := received["UserID"]
	if !ok {
		return false
	}
	uId := int64(raw.(float64))
	if !IsValidUserId(uId) {
		return false
	}
	return true
}
func IsValidUserId(userId int64) bool {
	_, ok := Users[userId]
	return ok
}

func AddUser(name string) (error, int64) {
	var u User
	var err error
	u.nickname = name
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
