package dto

import (
	"database/sql"
	"log"
)

type Exercise struct {
	ID          int64
	SplitID     int64
	Name        string
	Description string
	WeightFrom  float64
	WeightTo    float64
	RepsFrom    float64
	RepsTo      float64
	Sets        int64
}

func GetAllExercises(splitId int64, db *sql.DB) ([]Exercise, error) {
	rows, err := db.Query("SELECT ID, SplitID, Name, Description, WeightFrom, WeightTo, RepsFrom, RepsTo, Sets FROM exercises WHERE SplitID=?", splitId)
	if err != nil {
		log.Printf("GetAllExercises Error: %s", err.Error())
		return nil, err
	}

	exercises := []Exercise{}
	for rows.Next() {
		exercise := Exercise{}
		if err = rows.Scan(&exercise.ID, &exercise.SplitID, &exercise.Name, &exercise.Description, &exercise.WeightFrom, &exercise.WeightTo, &exercise.RepsFrom, &exercise.RepsTo, &exercise.Sets); err != nil {
			log.Printf("GetAllExercises Error: %s", err.Error())
			break
		}
		exercises = append(exercises, exercise)
	}

	return exercises, err
}

func GetExercise(exerciseId int64, db *sql.DB) (Exercise, error) {
	row := db.QueryRow("SELECT ID, SplitID, Name, Description, WeightFrom, WeightTo, RepsFrom, RepsTo, Sets FROM exercises WHERE ID=?", exerciseId)

	exercise := Exercise{}
	var err error
	if err = row.Scan(&exercise.ID, &exercise.SplitID, &exercise.Name, &exercise.Description, &exercise.WeightFrom, &exercise.WeightTo, &exercise.RepsFrom, &exercise.RepsTo, &exercise.Sets); err == sql.ErrNoRows {
		log.Printf("GetExercise Error: %s", err.Error())
		return Exercise{}, err
	}

	return exercise, nil
}

func GetAvailableExercises(splitId int64, workoutId int64, db *sql.DB) ([]Exercise, error) {
	rows, err := db.Query("SELECT * FROM exercises WHERE ID NOT IN (SELECT DISTINCT ExerciseID FROM workout_sets WHERE WorkoutID=?) AND SplitID=?", workoutId, splitId)
	if err != nil {
		log.Printf("GetAvailableExercises Error: %s", err.Error())
		return nil, err
	}

	exercises := []Exercise{}
	for rows.Next() {
		exercise := Exercise{}
		if err = rows.Scan(&exercise.ID, &exercise.SplitID, &exercise.Name, &exercise.Description, &exercise.WeightFrom, &exercise.WeightTo, &exercise.RepsFrom, &exercise.RepsTo, &exercise.Sets); err != nil {
			log.Printf("GetAvailableExercises Error: %s", err.Error())
			break
		}
		exercises = append(exercises, exercise)
	}

	return exercises, err
}
