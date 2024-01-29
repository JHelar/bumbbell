package dto

import (
	"database/sql"
	"log"
)

type Split struct {
	ID          int64
	UserID      int64
	Name        string
	Description string
}

func GetSplits(userId int64, db *sql.DB) ([]Split, error) {
	rows, err := db.Query("SELECT ID, UserID, Name, Description FROM splits WHERE UserID=?", userId)
	if err != nil {
		log.Printf("Error in GetSplits: %s", err.Error())
		return nil, err
	}

	splits := []Split{}

	for rows.Next() {
		split := Split{}
		if err = rows.Scan(&split.ID, &split.UserID, &split.Name, &split.Description); err != nil {
			log.Printf("Error in GetSplits: %s", err.Error())
			break
		}
		splits = append(splits, split)
	}

	return splits, err
}

func GetSplit(userId int64, splitId int64, db *sql.DB) (Split, error) {
	row := db.QueryRow(`
	SELECT ID, UserID, Name, Description FROM splits WHERE ID=? AND UserID=?
	`, splitId, userId)

	split := Split{}
	if err := row.Scan(&split.ID, &split.UserID, &split.Name, &split.Description); err != nil {
		log.Printf("Error in GetSplit: %s", err.Error())
		return Split{}, err
	}
	return split, nil
}

func CreateSplit(userId int64, name string, description string, db *sql.DB) (Split, error) {
	row := db.QueryRow(`
	INSERT INTO splits (UserID, Name, Description)
	VALUES (?, ?, ?)
	RETURNING ID, UserID, Name, Description
	`, userId, name, description)

	split := Split{}
	if err := row.Scan(&split.ID, &split.UserID, &split.Name, &split.Description); err != nil {
		log.Printf("Error in CreateSplit: %s", err.Error())
		return Split{}, err
	}
	return split, nil
}

func UpdateSplit(userId int64, splitId int64, name string, description string, db *sql.DB) (Split, error) {
	row := db.QueryRow(`
	UPDATE splits 
	SET Name=?,
		Description=?
	WHERE ID=? AND UserID=?
	RETURNING ID, UserID, Name, Description
	`, name, description, splitId, userId)

	split := Split{}
	if err := row.Scan(&split.ID, &split.UserID, &split.Name, &split.Description); err != nil {
		log.Printf("Error in Update split: %s", err.Error())
		return Split{}, err
	}
	return split, nil
}

func DeleteSplit(userId int64, splitId int64, db *sql.DB) error {
	deleteSplitResult, err := db.Exec(`
	DELETE FROM splits
	WHERE ID=? AND UserID=?
	`, splitId, userId)

	if err != nil {
		log.Printf("Error in DeleteSplit: %s", err.Error())
		return err
	}

	rows, err := deleteSplitResult.RowsAffected()

	if rows == 0 {
		return sql.ErrNoRows
	}

	return err
}
