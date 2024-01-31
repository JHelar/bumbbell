package server

import (
	"dumbbell/internal/model"
	"dumbbell/internal/templates"
	"net/http"
)

func (s *HttpServer) homeHandler(w http.ResponseWriter, r *http.Request) {
	viewModel := model.DashboardPageModel{
		Title:            "Dumbell",
		HasActiveWorkout: false,
	}

	activeWorkout, err := s.WorkoutService.GetActiveWorkout(TEST_USER_ID)
	if err == nil {
		viewModel.ActiveWorkout = activeWorkout
		viewModel.HasActiveWorkout = true
	}

	latestWorkoutSets, err := s.WorkoutService.GetLatestWorkoutSets(TEST_USER_ID)
	if err == nil {
		viewModel.LatestWorkoutSets = latestWorkoutSets
	}

	workoutActivity, err := s.WorkoutService.GetWorkoutActivity(TEST_USER_ID)
	if err == nil {
		viewModel.WorkoutActivity = workoutActivity
	}

	workoutSplits, err := s.WorkoutService.GetWorkoutSplits(TEST_USER_ID)
	if err == nil {
		viewModel.WorkoutSplits = workoutSplits
	}

	templates.ExecutePageTemplate(w, "index.html", viewModel)
}
