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

var ErrorSetLimitReached = errors.New("Set limit reached")
var ErrorWorkoutNotUpdated = errors.New("Workout not updated")

func NewWorkout(splitId int64, userId int64, db *sql.DB) (Workout, error) {
	row := db.QueryRow("INSERT INTO workouts (UserID, SplitID) VALUES (?,?) RETURNING ID, UserID, SplitID, StartedAt, CompletedAt", userId, splitId)
	workout := Workout{}

	var err error
	if err = row.Scan(&workout.ID, &workout.UserID, &workout.SplitID, &workout.StartedAt, &workout.CompletedAt); err != nil {
		log.Printf("NewWorkout Error: %s", err.Error())
	}
	return workout, err
}

func DeleteWorkout(userId int64, workoutId int64, db *sql.DB) error {
	result, err := db.Exec(`
	DELETE FROM workouts
	WHERE ID=? AND UserID=?
	`, workoutId, userId)

	if err != nil {
		log.Printf("Error deleting workout: %s", err.Error())
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error deleting workout: %s", err.Error())
		return err
	}

	if rows == 0 {
		log.Printf("Delete workout, no rows to delete: %s", err.Error())
		return sql.ErrNoRows
	}

	return nil
}

func GetActiveWorkout(userId int64, db *sql.DB) (Workout, error) {
	row := db.QueryRow("SELECT ID, UserID, SplitID, StartedAt, CompletedAt FROM workouts WHERE UserID=? AND CompletedAt IS NULL", userId)

	workout := Workout{}
	err := row.Scan(&workout.ID, &workout.UserID, &workout.SplitID, &workout.StartedAt, &workout.CompletedAt)

	return workout, err
}

func GetWorkout(userId int64, workoutId int64, db *sql.DB) (Workout, error) {
	row := db.QueryRow("SELECT ID, UserID, SplitID, StartedAt, CompletedAt FROM workouts WHERE ID=? AND UserID=?", workoutId, userId)

	var err error
	workout := Workout{}
	if err = row.Scan(&workout.ID, &workout.UserID, &workout.SplitID, &workout.StartedAt, &workout.CompletedAt); err != nil {
		if err != sql.ErrNoRows {
			log.Printf("GetWorkout Error: %s", err.Error())
		}
	}

	return workout, err
}

func GetAllCompletedWorkouts(userId int64, db *sql.DB) ([]Workout, error) {
	rows, err := db.Query(`
	SELECT ID, UserID, SplitID, StartedAt, CompletedAt FROM workouts
	WHERE UserID=? AND CompletedAt IS NOT NULL
	ORDER BY StartedAt
	`, userId)

	if err != nil {
		log.Printf("GetAllCompletedWorkouts error: %s", err.Error())
		return nil, err
	}

	workouts := []Workout{}
	for rows.Next() {
		workout := Workout{}
		if err = rows.Scan(&workout.ID, &workout.UserID, &workout.SplitID, &workout.StartedAt, &workout.CompletedAt); err != nil {
			break
		}

		workouts = append(workouts, workout)
	}

	return workouts, err
}

func GetAllCompletedWorkoutsForSplit(userId int64, splitId int64, db *sql.DB) ([]Workout, error) {
	rows, err := db.Query(`
	SELECT ID, UserID, SplitID, StartedAt, CompletedAt FROM workouts
	WHERE UserID=? AND SplitID=? AND CompletedAt IS NOT NULL
	ORDER BY StartedAt
	`, userId, splitId)

	if err != nil {
		log.Printf("GetAllCompletedWorkoutsForSplit error: %s", err.Error())
		return nil, err
	}

	workouts := []Workout{}
	for rows.Next() {
		workout := Workout{}
		if err = rows.Scan(&workout.ID, &workout.UserID, &workout.SplitID, &workout.StartedAt, &workout.CompletedAt); err != nil {
			break
		}

		workouts = append(workouts, workout)
	}

	return workouts, err
}

func CompleteWorkout(workoutId int64, db *sql.DB) error {
	result, err := db.Exec(`
	UPDATE workouts
	SET CompletedAt=CURRENT_TIMESTAMP
	WHERE ID=? AND CompletedAt IS NULL
	`, workoutId)

	if rows, _ := result.RowsAffected(); rows == 0 {
		return ErrorWorkoutNotUpdated
	}

	return err
}

func CreateNewSet(workoutId int64, exerciseId int64, db *sql.DB) (WorkoutSet, error) {
	_, getActiveWorkoutSetErr := GetActiveWorkoutSet(workoutId, db)
	if getActiveWorkoutSetErr != sql.ErrNoRows {
		return WorkoutSet{}, getActiveWorkoutSetErr
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

func CreateNextSet(workoutId int64, rating SetStatus, db *sql.DB) (WorkoutSet, error) {
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
		return WorkoutSet{}, ErrorSetLimitReached
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
		log.Printf("GetActiveWorkoutSet Error: %s", err.Error())
		return WorkoutSet{}, err
	}

	exercise, err := GetExercise(workoutSet.ExerciseID, db)
	if err != nil {
		return WorkoutSet{}, err
	}

	workoutSet.Sets = exercise.Sets
	return workoutSet, err
}

func GetAllWorkoutSets(userId int64, limit int, db *sql.DB) ([]WorkoutSet, error) {
	rows, err := db.Query(`
	SELECT SetNumber, WorkoutID, ExerciseID, StartedAt, CompletedAt, SetRating, WeightFrom, WeightTo, RepsFrom, RepsTo 
	FROM workout_sets 
	WHERE WorkoutID IN (
		SELECT ID FROM workouts
		WHERE UserID=?
	)
	ORDER BY CompletedAt DESC NULLS FIRST
	LIMIT ?
	`, userId, limit)

	if err != nil {
		log.Printf("GetAllWorkoutSets error: %s", err.Error())
		return nil, err
	}

	workoutSets := []WorkoutSet{}
	for rows.Next() {
		workoutSet := WorkoutSet{}
		if err = rows.Scan(&workoutSet.SetNumber, &workoutSet.WorkoutID, &workoutSet.ExerciseID, &workoutSet.StartedAt, &workoutSet.CompletedAt, &workoutSet.SetRating, &workoutSet.WeightTo, &workoutSet.WeightFrom, &workoutSet.RepsFrom, &workoutSet.RepsTo); err != nil {
			log.Printf("GetAllWorkoutSets Error: %s", err.Error())
			break
		}
		workoutSets = append(workoutSets, workoutSet)
	}

	return workoutSets, err
}

func GetAllWorkoutSetsForExercise(exerciseId int64, db *sql.DB) ([]WorkoutSet, error) {
	rows, err := db.Query(`
	SELECT SetNumber, WorkoutID, ExerciseID, StartedAt, CompletedAt, SetRating, WeightFrom, WeightTo, RepsFrom, RepsTo 
	FROM workout_sets 
	WHERE ExerciseID=?
	ORDER BY CompletedAt DESC NULLS FIRST
	`, exerciseId)

	if err != nil {
		log.Printf("GetAllWorkoutSetsForExercise error: %s", err.Error())
		return nil, err
	}

	workoutSets := []WorkoutSet{}
	for rows.Next() {
		workoutSet := WorkoutSet{}
		if err = rows.Scan(&workoutSet.SetNumber, &workoutSet.WorkoutID, &workoutSet.ExerciseID, &workoutSet.StartedAt, &workoutSet.CompletedAt, &workoutSet.SetRating, &workoutSet.WeightTo, &workoutSet.WeightFrom, &workoutSet.RepsFrom, &workoutSet.RepsTo); err != nil {
			log.Printf("GetAllWorkoutSetsForExercise Error: %s", err.Error())
			break
		}
		workoutSets = append(workoutSets, workoutSet)
	}

	return workoutSets, err
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
