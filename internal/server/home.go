package server

import (
	"dumbbell/internal/templates"
	"net/http"
)

func (s *HttpServer) dashboardHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecutePageTemplate(w, "index.html", map[string]interface{}{
		"Title": "Dumbell",
	})
}
