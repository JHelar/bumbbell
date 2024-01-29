package server

import (
	"dumbbell/internal/templates"
	"net/http"
)

func (s *HttpServer) workoutsHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecutePageTemplate(w, "workouts.html", map[string]interface{}{
		"Title": "Dumbbell - Workouts",
	})
}
