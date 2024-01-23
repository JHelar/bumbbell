package model

import (
	"dumbbell/internal/dto"
)

type ExerciseViewModel struct {
	Name        string
	Description string
	ImageSrc    string
	WeightFrom  float64
	WeightTo    float64
	RepsFrom    float64
	RepsTo      float64
	Sets        []dto.SetStatus
}

type CardViewModel struct {
	ID          string
	Name        string
	Description string
}

type EditWorkoutTableExerciseModel struct {
	ID          int64
	Name        string
	Description string
	WeightFrom  float64
	WeightTo    float64
	RepsFrom    float64
	RepsTo      float64
	ImageSrc    string
	Sets        int64
}

type EditWorkoutTableSplitModel struct {
	ID          int64
	Name        string
	Description string
	Exercises   []EditWorkoutTableExerciseModel
}

type UserSettingsModel struct {
	Title  string
	Splits []EditWorkoutTableSplitModel
}
