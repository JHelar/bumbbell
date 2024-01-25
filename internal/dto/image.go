package dto

import (
	"database/sql"
	"log"
)

type ImageType string

const (
	ImageTypeJpeg ImageType = "jpeg"
)

type Image struct {
	ID          int64
	Content     []byte
	ContentType ImageType
}

func CreateImage(contentType ImageType, content []byte, db *sql.DB) (Image, error) {
	row := db.QueryRow(
		`INSERT INTO images (ContentType, Content)
		VALUES (?,?)
		RETURNING ID`,
		contentType, content)

	var err error
	image := Image{
		Content:     content,
		ContentType: contentType,
	}
	if err = row.Scan(&image.ID); err != nil {
		log.Printf("NewImage error: %s", err.Error())
		return Image{}, err
	}

	return image, nil
}

func DeleteImage(imageId int64, db *sql.DB) error {
	var err error
	res, err := db.Exec(`
	DELETE FROM images
	WHERE ID=?
	`, imageId)

	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}

	return err
}
