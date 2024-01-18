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
