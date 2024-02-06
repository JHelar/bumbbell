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

	latestWorkoutSets, err := s.WorkoutService.GetLatestWorkoutSets(userId)
	if err == nil {
		viewModel.LatestWorkoutSets = latestWorkoutSets
	}

	workoutActivity, err := s.WorkoutService.GetWorkoutActivity(userId)
	if err == nil {
		viewModel.WorkoutActivity = workoutActivity
	}

	workoutSplits, err := s.WorkoutService.GetWorkoutSplits(userId)
	if err == nil {
		viewModel.WorkoutSplits = workoutSplits
	}

	w.Header().Add("HX-Reselect", "#container")
	w.Header().Add("HX-Retarget", "#container")
	w.Header().Add("HX-Reswap", "outerHTML")
	templates.ExecutePageTemplate(w, "index.html", viewModel)
}
