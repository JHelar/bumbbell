package dto

import (
	"database/sql"
	"time"
)

type SetStatus string

const (
	SetGood        SetStatus = "good"
	SetBad         SetStatus = "bad"
	SetUncompleted SetStatus = "uncompleted"
	SetCurrent     SetStatus = "current"
)

type Exercise struct {
	ID          string
	SplitID     string
	Name        string
	Description string
	WeightFrom  float64
	WeightTo    float64
	RepsFrom    float64
	RepsTo      float64
	Sets        int
}

type Split struct {
	ID     string
	UserID string
	Name   string
}

type Workout struct {
	ID          string
	UserID      string
	SplitID     string
	StartedAt   time.Time
	CompletedAt sql.NullTime
}

type WorkoutExercise struct {
	ID          string
	WorkoutID   string
	ExerciseID  string
	StartedAt   time.Time
	CompletedAt sql.NullTime
}

type WorkoutSet struct {
	WorkoutExerciseID string
	StartedAt         time.Time
	CompletedAt       sql.NullTime
	Status            SetStatus
	SetNumber         int
}

type User struct {
	ID string
}
