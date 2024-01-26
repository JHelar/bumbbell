package dto

import (
	"database/sql"
	"errors"
	"fmt"
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

func (e *Exercise) GetImageURL() string {
	return fmt.Sprintf("/exercise/image/%d", e.ID)
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

func UpdateExercise(
	id int64,
	name string,
	imageId *int64,
	description string,
	weightFrom float64,
	weightTo float64,
	repsFrom int64,
	repsTo int64,
	sets int64, db *sql.DB) (Exercise, error) {

	var row *sql.Row
	var err error

	if imageId == nil {
		row = db.QueryRow(`
		UPDATE exercises 
		SET Name=?,
		Description=?,
		WeightFrom=?,
		WeightTo=?,
		RepsFrom=?,
		RepsTo=?,
		Sets=?
		WHERE ID=?
		RETURNING ID, SplitID, Name, Description, WeightFrom, WeightTo, RepsFrom, RepsTo, Sets
		`, name, description, weightFrom, weightTo, repsFrom, repsTo, sets, id)
	} else {
		row = db.QueryRow(`
		UPDATE exercises 
		SET Name=?,
		Description=?,
		WeightFrom=?,
		WeightTo=?,
		RepsFrom=?,
		RepsTo=?,
		Sets=?
		WHERE ID=?
		RETURNING ID, SplitID, Name, Description, WeightFrom, WeightTo, RepsFrom, RepsTo, Sets
		`, name, description, weightFrom, weightTo, repsFrom, repsTo, sets, id)

		err = ReplaceExerciseImage(id, *imageId, db)
	}

	if err != nil {
		log.Printf("UpdateExercise Error: %s", err.Error())
		return Exercise{}, err
	}

	exercise := Exercise{}
	if err = row.Scan(&exercise.ID, &exercise.SplitID, &exercise.Name, &exercise.Description, &exercise.WeightFrom, &exercise.WeightTo, &exercise.RepsFrom, &exercise.RepsTo, &exercise.Sets); err == sql.ErrNoRows {
		log.Printf("UpdateExercise Error: %s", err.Error())
		return Exercise{}, err
	}

	return exercise, err
}

func CreateExercise(
	splitId int64,
	imageId *int64,
	name string,
	description string,
	weightFrom float64,
	weightTo float64,
	repsFrom int64,
	repsTo int64,
	sets int64,
	db *sql.DB) (Exercise, error) {

	if imageId == nil {
		return Exercise{}, errors.New("No image provided")
	}

	row := db.QueryRow(`INSERT INTO exercises (SplitID, Name, Description, ImageID, WeightFrom, WeightTo, RepsFrom, RepsTo, Sets)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	RETURNING ID, SplitID, Name, Description, WeightFrom, WeightTo, RepsFrom, RepsTo, Sets
	`, splitId,
		name,
		description,
		imageId,
		weightFrom,
		weightTo,
		repsFrom,
		repsTo,
		sets)

	exercise := Exercise{}
	var err error
	if err = row.Scan(&exercise.ID, &exercise.SplitID, &exercise.Name, &exercise.Description, &exercise.WeightFrom, &exercise.WeightTo, &exercise.RepsFrom, &exercise.RepsTo, &exercise.Sets); err != nil {
		log.Printf("CreateExercise Error: %s", err.Error())
		return Exercise{}, err
	}

	return exercise, err
}

func DeleteExercise(exerciseId int64, db *sql.DB) error {
	var err error

	result, err := db.Exec(`
	DELETE FROM exercises
		WHERE ID=?
	`, exerciseId)

	rows, err := result.RowsAffected()

	if rows == 0 {
		return sql.ErrNoRows
	}

	return err
}

func GetAvailableExercises(splitId int64, workoutId int64, db *sql.DB) ([]Exercise, error) {
	rows, err := db.Query(`
	SELECT ID, SplitID, Name, Description, WeightFrom, WeightTo, RepsFrom, RepsTo, Sets FROM exercises 
	WHERE ID NOT IN (
		SELECT DISTINCT ExerciseID 
		FROM workout_sets 
		WHERE WorkoutID=?
	) 
	AND SplitID=?
	`, workoutId, splitId)
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

func ReplaceExerciseImage(exerciseId int64, newImageId int64, db *sql.DB) error {
	var err error
	selectRow := db.QueryRow(`
	SELECT ImageID FROM exercises
	WHERE ID=?
	`, exerciseId)

	var oldImageId int64
	if err := selectRow.Scan(&oldImageId); err != nil {
		return err
	}

	updateResult, err := db.Exec(`
	UPDATE exercises
	SET ImageID=?
	WHERE ID=?
	`, newImageId, exerciseId)

	if err != nil {
		return err
	}

	updateCount, err := updateResult.RowsAffected()
	if updateCount == 0 {
		return sql.ErrNoRows
	}

	if err != nil {
		return err
	}

	err = DeleteImage(oldImageId, db)

	return err
}

func GetExerciseImage(exerciseId int64, db *sql.DB) (Image, error) {
	row := db.QueryRow(`
		SELECT Content, ContentType FROM images
		WHERE ID=(
			SELECT ImageID FROM exercises
			WHERE ID=?
		)
	`, exerciseId)

	image := Image{}
	err := row.Scan(&image.Content, &image.ContentType)

	return image, err
}
