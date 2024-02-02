package service

import (
	"database/sql"
)

type ExerciseService struct {
	DB *sql.DB
}

func NewExerciseService(db *sql.DB) ExerciseService {
	return ExerciseService{DB: db}
}
