package server

import (
	"dumbbell/internal/dto"
	"dumbbell/internal/model"
	"dumbbell/internal/templates"
	"net/http"
)

func (s *HttpServer) homeHandler(w http.ResponseWriter, r *http.Request) {
	viewModel := model.DashboardPageModel{
		Title:            "Dumbell",
		HasActiveWorkout: false,
	}

	activeWorkout, err := dto.GetActiveWorkout(TEST_USER_ID, s.DB)
	if err == nil {
		activeWorkoutData, _ := s.WorkoutService.GetActiveWorkoutData(TEST_USER_ID, activeWorkout.ID)
		viewModel.ActiveWorkout = activeWorkoutData
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
