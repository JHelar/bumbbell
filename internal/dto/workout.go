package dto

import (
	"database/sql"
	"errors"
	"log"
	"time"
)

type SetStatus string

const (
	SetGood        SetStatus = "good"
	SetBad         SetStatus = "bad"
	SetUncompleted SetStatus = "uncompleted"
	SetCurrent     SetStatus = "current"
)

type Workout struct {
	ID          int64
	UserID      int64
	SplitID     int64
	StartedAt   time.Time
	CompletedAt sql.NullTime
}

type WorkoutSet struct {
	SetNumber   int64
	WorkoutID   int64
	ExerciseID  int64
	StartedAt   time.Time
	CompletedAt sql.NullTime
	SetRating   SetStatus
	WeightFrom  float64
	WeightTo    float64
	RepsFrom    float64
	RepsTo      float64
	Sets        int64
}

func NewWorkout(splitId int64, userId int64, db *sql.DB) (Workout, error) {
	activeWorkout, err := GetActiveWorkout(userId, db)
	if err == nil {
		return activeWorkout, nil
	}

	row := db.QueryRow("INSERT INTO workouts (UserID, SplitID) VALUES (?,?) RETURNING ID, UserID, SplitID, StartedAt, CompletedAt", userId, splitId)
	workout := Workout{}
	if err = row.Scan(&workout.ID, &workout.UserID, &workout.SplitID, &workout.StartedAt, &workout.CompletedAt); err != nil {
		log.Printf("NewWorkout Error: %s", err.Error())
	}
	return workout, err
}

func GetActiveWorkout(userId int64, db *sql.DB) (Workout, error) {
	row := db.QueryRow("SELECT ID, UserID, SplitID, StartedAt, CompletedAt FROM workouts WHERE UserID=? AND CompletedAt IS NULL", userId)

	var err error
	workout := Workout{}
	if err = row.Scan(&workout.ID, &workout.UserID, &workout.SplitID, &workout.StartedAt, &workout.CompletedAt); err != nil {
		if err != sql.ErrNoRows {
			log.Printf("GetActiveWorkout Error: %s", err.Error())
		}
	}
	return workout, err
}

func NewSet(workoutId int64, exerciseId int64, db *sql.DB) (WorkoutSet, error) {
	_, err := GetActiveWorkoutSet(workoutId, db)
	if err == nil {
		return WorkoutSet{}, errors.New("NewSet Error a set is allready active")
	}

	exercise, err := GetExercise(exerciseId, db)
	if err != nil {
		return WorkoutSet{}, err
	}

	setNumber := int64(1)

	if setNumber > exercise.Sets {
		log.Print("NextSet Error set limit reached")
		return WorkoutSet{}, errors.New("Set limit reached")
	}

	row := db.QueryRow(`
		INSERT INTO workout_sets (SetNumber,WorkoutID,ExerciseID,WeightFrom,WeightTo,RepsFrom,RepsTo)
		VALUES (?,?,?,?,?,?,?)
		RETURNING StartedAt, StartedAt, SetRating
	`, setNumber, workoutId, exerciseId, exercise.WeightFrom, exercise.WeightTo, exercise.RepsFrom, exercise.RepsTo)

	workoutSet := WorkoutSet{
		SetNumber:  setNumber,
		WorkoutID:  workoutId,
		ExerciseID: exerciseId,
		WeightFrom: exercise.WeightFrom,
		WeightTo:   exercise.WeightTo,
		RepsFrom:   exercise.RepsFrom,
		RepsTo:     exercise.RepsTo,
		Sets:       exercise.Sets,
	}

	if err = row.Scan(&workoutSet.StartedAt, &workoutSet.CompletedAt, &workoutSet.SetRating); err != nil {
		log.Printf("NewSet Error: %s", err.Error())
		return WorkoutSet{}, nil
	}

	return workoutSet, nil
}

func NextSet(workoutId int64, rating SetStatus, db *sql.DB) (WorkoutSet, error) {
	activeWorkoutSet, err := UpdateActiveWorkoutSet(workoutId, rating, db)
	if err != nil {
		return WorkoutSet{}, err
	}

	exercise, err := GetExercise(activeWorkoutSet.ExerciseID, db)
	if err != nil {
		return WorkoutSet{}, err
	}

	setNumber := activeWorkoutSet.SetNumber + 1
	if setNumber > exercise.Sets {
		log.Print("NextSet Error set limit reached")
		return WorkoutSet{}, errors.New("Set limit reached")
	}

	row := db.QueryRow(`
		INSERT INTO workout_sets (SetNumber,WorkoutID,ExerciseID,WeightFrom,WeightTo,RepsFrom,RepsTo)
		VALUES (?,?,?,?,?,?,?)
		RETURNING StartedAt, StartedAt, SetRating
	`, setNumber, workoutId, activeWorkoutSet.ExerciseID, exercise.WeightFrom, exercise.WeightTo, exercise.RepsFrom, exercise.RepsTo)

	workoutSet := WorkoutSet{
		SetNumber:  setNumber,
		WorkoutID:  workoutId,
		ExerciseID: activeWorkoutSet.ExerciseID,
		WeightFrom: exercise.WeightFrom,
		WeightTo:   exercise.WeightTo,
		RepsFrom:   exercise.RepsFrom,
		RepsTo:     exercise.RepsTo,
		Sets:       exercise.Sets,
	}

	if err = row.Scan(&workoutSet.StartedAt, &workoutSet.CompletedAt, &workoutSet.SetRating); err != nil {
		log.Printf("NewSet Error: %s", err.Error())
		return WorkoutSet{}, nil
	}

	return workoutSet, nil
}

func GetCompletedWorkoutSets(workoutId int64, exerciseId int64, db *sql.DB) ([]WorkoutSet, error) {
	rows, err := db.Query("SELECT SetNumber, WorkoutID, ExerciseID, StartedAt, CompletedAt, SetRating, WeightFrom, WeightTo, RepsFrom, RepsTo FROM workout_sets WHERE WorkoutID=? AND ExerciseID=? AND CompletedAt IS NOT NULL ORDER BY SetNumber ASC", workoutId, exerciseId)
	if err != nil {
		log.Printf("GetCompletedWorkoutSets Error: %s", err.Error())
		return nil, err
	}

	workoutSets := []WorkoutSet{}
	for rows.Next() {
		workoutSet := WorkoutSet{}
		if err = rows.Scan(&workoutSet.SetNumber, &workoutSet.WorkoutID, &workoutSet.ExerciseID, &workoutSet.StartedAt, &workoutSet.CompletedAt, &workoutSet.SetRating, &workoutSet.WeightTo, &workoutSet.WeightFrom, &workoutSet.RepsFrom, &workoutSet.RepsTo); err != nil {
			log.Printf("GetCompletedWorkoutSets Error: %s", err.Error())
			break
		}
		workoutSets = append(workoutSets, workoutSet)
	}

	return workoutSets, err
}

func GetActiveWorkoutSet(workoutId int64, db *sql.DB) (WorkoutSet, error) {
	row := db.QueryRow("SELECT SetNumber, WorkoutID, ExerciseID, StartedAt, CompletedAt, SetRating, WeightFrom, WeightTo, RepsFrom, RepsTo FROM workout_sets WHERE WorkoutID=? AND CompletedAt IS NULL", workoutId)

	var err error
	workoutSet := WorkoutSet{}
	if err = row.Scan(&workoutSet.SetNumber, &workoutSet.WorkoutID, &workoutSet.ExerciseID, &workoutSet.StartedAt, &workoutSet.CompletedAt, &workoutSet.SetRating, &workoutSet.WeightTo, &workoutSet.WeightFrom, &workoutSet.RepsFrom, &workoutSet.RepsTo); err != nil {
		if err != sql.ErrNoRows {
			log.Printf("GetActiveWorkoutSet Error: %s", err.Error())
		}
	}

	exercise, err := GetExercise(workoutSet.ExerciseID, db)
	if err != nil {
		return WorkoutSet{}, err
	}

	workoutSet.Sets = exercise.Sets
	return workoutSet, err
}

func UpdateActiveWorkoutSet(workoutId int64, rating SetStatus, db *sql.DB) (WorkoutSet, error) {
	row := db.QueryRow(`
		UPDATE workout_sets
		SET CompletedAt=CURRENT_TIMESTAMP, SetRating=?
		WHERE WorkoutID=? AND CompletedAt IS NULL
		RETURNING SetNumber, WorkoutID, ExerciseID
		`, rating, workoutId)

	var err error
	workoutSet := WorkoutSet{}
	if err = row.Scan(&workoutSet.SetNumber, &workoutSet.WorkoutID, &workoutSet.ExerciseID); err != nil {
		if err != sql.ErrNoRows {
			log.Printf("UpdateActiveWorkoutSet Error: %s", err.Error())
		}
	}
	return workoutSet, err
}
