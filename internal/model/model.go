package model

import (
	"dumbbell/internal/dto"
)

type ExerciseViewModel struct {
	Name        string
	Description string
	ImageSrc    string
	WeightFrom  float32
	WeightTo    float32
	RepsFrom    float32
	RepsTo      float32
	Sets        []dto.SetStatus
}

type CardViewModel struct {
	Name        string
	Description string
}
