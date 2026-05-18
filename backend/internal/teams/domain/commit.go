package domain

import "time"

type CommitInput struct {
	ProjectID    string
	SHA          string
	Message      string
	Branch       string
	FilesChanged int
	LinesAdded   int
	LinesRemoved int
	AILinesPct   *int
	AuthoredAt   time.Time
}
