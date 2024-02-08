package service

import (
	"database/sql"
	"dumbbell/internal/dto"
	"dumbbell/internal/model"
	"errors"
	"log"
	"net/http"

	"github.com/michaeljs1990/sqlitestore"

	"golang.org/x/crypto/bcrypt"
)

var SESSION_COOKIE_NAME = "user-session"

var InvalidCredentialsError = errors.New("Invalid credentials")

type SessionService struct {
	DB    *sql.DB
	Store *sqlitestore.SqliteStore
}

func NewSessionService(db *sql.DB) SessionService {
	store, err := sqlitestore.NewSqliteStoreFromConnection(db, "user_sessions", "/", 0, []byte("NOT_SO_SECRET_KEY"))
	if err != nil {
		panic(err)
	}

	return SessionService{
		DB:    db,
		Store: store,
	}
}

type LoginUserData struct {
	Email    string
	Password string
	Remember bool
}

func (s *SessionService) LoginUser(w http.ResponseWriter, r *http.Request, data LoginUserData) error {
	user, getUserErr := dto.GetUserByEmail(data.Email, s.DB)
	if getUserErr != nil {
		if getUserErr == sql.ErrNoRows {
			return InvalidCredentialsError
		}
		return getUserErr
	}

	authUserErr := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(data.Password))
	if authUserErr != nil {
		return InvalidCredentialsError
	}

	session, getSessionErr := s.Store.Get(r, SESSION_COOKIE_NAME)
	if getSessionErr != nil {
		return getSessionErr
	}

	if data.Remember {
		session.Options.MaxAge = 3600
	} else {
		session.Options.MaxAge = 0
	}

	session.Values["user_id"] = user.ID
	saveSessionErr := session.Save(r, w)
	if saveSessionErr != nil {
		return saveSessionErr
	}

	return nil
}

func (s *SessionService) LogoutUser(w http.ResponseWriter, r *http.Request) error {
	session, err := s.Store.Get(r, SESSION_COOKIE_NAME)
	if err != nil {
		return err
	}

	session.Options.MaxAge = -1
	err = session.Save(r, w)
	if err != nil {
		return err
	}

	return nil
}

func (s *SessionService) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := s.Store.Get(r, SESSION_COOKIE_NAME)
		if err != nil {
			log.Printf("Error getting session: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if session.Values["user_id"] == nil {
			log.Printf("User not authenticated")
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *SessionService) GetUserId(r *http.Request) (int64, error) {
	session, err := s.Store.Get(r, SESSION_COOKIE_NAME)
	if err != nil {
		return 0, err
	}

	userId := session.Values["user_id"]
	if userId == nil {
		return 0, nil
	}

	return userId.(int64), nil
}

func (s *SessionService) MustGetUserId(w http.ResponseWriter, r *http.Request) int64 {
	userId, err := s.GetUserId(r)
	if err != nil {
		log.Printf("Error getting user id: %s", err.Error())
		w.Header().Add("HX-Replace-Url", "/login")
		http.Redirect(w, r, "/login", http.StatusFound)
	}

	return userId
}

func (s *SessionService) IsAuthenticated(r *http.Request) bool {
	session, err := s.Store.Get(r, SESSION_COOKIE_NAME)
	if err != nil {
		return false
	}

	return session.Values["user_id"] != nil
}

func (s *SessionService) GetHeaderModel(r *http.Request) model.HeaderModel {
	session, err := s.Store.Get(r, SESSION_COOKIE_NAME)
	if err != nil {
		return model.HeaderModel{
			IsLoggedIn: false,
		}
	}

	userId := session.Values["user_id"]
	if userId == nil {
		return model.HeaderModel{
			IsLoggedIn: false,
		}
	}

	user, err := dto.GetUserById(userId.(int64), s.DB)
	if err != nil {
		return model.HeaderModel{
			IsLoggedIn: false,
		}
	}

	return model.HeaderModel{
		IsLoggedIn:   true,
		UserEmail:    user.Email,
		UserImageSrc: user.GetImageURL(),
	}
}
