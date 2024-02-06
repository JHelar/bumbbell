package dto

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           int64
	Email        string
	PasswordHash []byte
}

func GetUserByEmail(email string, db *sql.DB) (User, error) {
	row := db.QueryRow(`
	SELECT ID, Email, PasswordHash FROM users WHERE Email=?
	`, email)

	user := User{}
	if err := row.Scan(&user.ID, &user.Email, &user.PasswordHash); err != nil {
		return User{}, err
	}
	return user, nil
}

func GetUserById(id int64, db *sql.DB) (User, error) {
	row := db.QueryRow(`
	SELECT ID, Email FROM users WHERE ID=?
	`, id)

	user := User{}
	if err := row.Scan(&user.ID, &user.Email); err != nil {
		return User{}, err
	}
	return user, nil
}

func CreateUser(email, password string, db *sql.DB) (User, error) {
	passwordHash, generatePasswordErr := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if generatePasswordErr != nil {
		return User{}, generatePasswordErr
	}

	row := db.QueryRow(`
	INSERT INTO users (Email, PasswordHash)
	VALUES (?, ?)
	RETURNING ID
	`, email, passwordHash)

	user := User{
		Email: email,
	}
	if err := row.Scan(&user.ID); err != nil {
		return User{}, err
	}

	return user, nil
}

func (u *User) GetImageURL() string {
	normalizedEmail := strings.ToLower(strings.Trim(u.Email, " "))
	sha256 := sha256.New()
	sha256.Write([]byte(normalizedEmail))
	sha256String := hex.EncodeToString(sha256.Sum(nil))

	return fmt.Sprintf("https://gravatar.com/avatar/%s", sha256String)
}
