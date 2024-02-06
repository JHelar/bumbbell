package server

import (
	"dumbbell/internal/dto"
	"dumbbell/internal/model"
	"dumbbell/internal/service"
	"dumbbell/internal/templates"
	"log"
	"net/http"
)

func (s *HttpServer) RegisterUser(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirm-password")

	if password != confirmPassword {
		log.Printf("Passwords do not match")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, err := dto.CreateUser(email, password, s.DB)

	if err != nil {
		log.Printf("Error creating user: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	loginErr := s.SessionService.LoginUser(w, r, service.LoginUserData{
		Email:    email,
		Password: password,
		Remember: true,
	})

	if loginErr != nil {
		log.Printf("Error logging in user: %s", loginErr.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("HX-Replace-Url", "/")
	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *HttpServer) LoginUser(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")

	loginErr := s.SessionService.LoginUser(w, r, service.LoginUserData{
		Email:    email,
		Password: password,
		Remember: true,
	})
	if loginErr != nil {
		log.Printf("Error logging in user: %s", loginErr.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("HX-Replace-Url", "/")
	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *HttpServer) LogoutUser(w http.ResponseWriter, r *http.Request) {
	logoutErr := s.SessionService.LogoutUser(w, r)
	if logoutErr != nil {
		log.Printf("Error logging out user: %s", logoutErr.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("HX-Replace-Url", "/login")
	http.Redirect(w, r, "/login", http.StatusFound)
}

func (s *HttpServer) ResetPassword(w http.ResponseWriter, r *http.Request) {
	// ...
}

func (s *HttpServer) ChangePassword(w http.ResponseWriter, r *http.Request) {
	// ...
}

func (s *HttpServer) loginPageHandler(w http.ResponseWriter, r *http.Request) {
	if s.SessionService.IsAuthenticated(r) {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	templates.ExecutePageTemplate(w, "login.html", model.LoginPageModel{
		Title:  "Dumbbell - Login",
		Header: s.SessionService.GetHeaderModel(r),
	})
}

func (s *HttpServer) signupPageHandler(w http.ResponseWriter, r *http.Request) {
	if s.SessionService.IsAuthenticated(r) {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	templates.ExecutePageTemplate(w, "signup.html", model.LoginPageModel{
		Title:  "Dumbbell - Signup",
		Header: s.SessionService.GetHeaderModel(r),
	})
}
