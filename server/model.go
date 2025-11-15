package server

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
	header     ProblemHeader
	difficulty ProblemDifficulty
	id         uint16
	objective  string
}
