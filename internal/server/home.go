package server

import (
	"database/sql"
	"dumbbell/internal/dto"
	"dumbbell/internal/model"
	"dumbbell/internal/service"
	"dumbbell/internal/templates"
	"log"
	"net/http"
)

func (s *HttpServer) homeHandler(w http.ResponseWriter, r *http.Request) {
	if !s.SessionService.IsAuthenticated(r) {
		w.Header().Add("HX-Replace-Url", "/login")
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	userId := s.SessionService.MustGetUserId(w, r)
	viewModel := model.DashboardPageModel{
		Title:            "Dumbell",
		HasActiveWorkout: false,
		Splits:           nil,
		Header:           s.SessionService.GetHeaderModel(r),
	}

	activeWorkout, err := dto.GetActiveWorkout(userId, s.DB)
	if err == nil {
		activeWorkoutData, _ := s.WorkoutService.GetActiveWorkoutData(userId, activeWorkout.ID)
		workoutMetadata := service.GetWorkoutMetaData(activeWorkout)
		viewModel.ActiveWorkout = activeWorkoutData
		viewModel.HasActiveWorkout = true
		viewModel.WorkoutStart = workoutMetadata.WorkoutStart
		viewModel.WorkoutDuration = workoutMetadata.WorkoutDuration
		viewModel.WorkoutStartedAt = workoutMetadata.WorkoutStartedAt
	} else if err == sql.ErrNoRows {
		splits, err := s.WorkoutService.GetSplitCards(userId)
		if err == nil {
			viewModel.Splits = splits
		} else {
			log.Print("Error getting splits: ", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else if err != nil {
		log.Print("Error getting active workout: ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	latestWorkoutSets, latestWorkoutErr := s.WorkoutService.GetLatestWorkoutSets(userId)
	viewModel.LatestWorkoutSets = latestWorkoutSets
	if latestWorkoutErr != nil {
		log.Print("Error getting latest workout sets: ", latestWorkoutErr.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	workoutActivity, getWorkoutActivityErr := s.WorkoutService.GetWorkoutActivity(userId)
	viewModel.WorkoutActivity = workoutActivity
	if getWorkoutActivityErr != nil {
		log.Printf("Error getting workout activity: %s", getWorkoutActivityErr.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	workoutSplits, getWorkoutSplitsErr := s.WorkoutService.GetWorkoutSplits(userId)
	viewModel.WorkoutSplits = workoutSplits
	if getWorkoutSplitsErr != nil {
		log.Printf("Error getting workout splits: %s", getWorkoutSplitsErr.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if s.HtmxService.IsHtmxRequest(r) {
		exercuseTemplateErr := templates.Home.Execute(w, viewModel)
		if exercuseTemplateErr != nil {
			log.Print("Error home template: ", exercuseTemplateErr.Error())
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	templates.ExecutePageTemplate(w, "index.html", viewModel)
}
