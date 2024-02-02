package server

import (
	"database/sql"
	"dumbbell/internal/dto"
	"dumbbell/internal/model"
	"dumbbell/internal/templates"
	"log"
	"net/http"
)

func (s *HttpServer) homeHandler(w http.ResponseWriter, r *http.Request) {
	viewModel := model.DashboardPageModel{
		Title:            "Dumbell",
		HasActiveWorkout: false,
		Splits:           nil,
	}

	activeWorkout, err := dto.GetActiveWorkout(TEST_USER_ID, s.DB)
	if err == nil {
		activeWorkoutData, _ := s.WorkoutService.GetActiveWorkoutData(TEST_USER_ID, activeWorkout.ID)
		viewModel.ActiveWorkout = activeWorkoutData
		viewModel.HasActiveWorkout = true
	} else if err == sql.ErrNoRows {
		splits, err := s.WorkoutService.GetSplitCards(TEST_USER_ID)
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

	w.Header().Add("HX-Reselect", "#container")
	w.Header().Add("HX-Retarget", "#container")
	w.Header().Add("HX-Reswap", "outerHTML")
	templates.ExecutePageTemplate(w, "index.html", viewModel)
}
