package server

import (
	"dumbbell/internal/dto"
	"dumbbell/internal/model"
	"dumbbell/internal/service"
	"dumbbell/internal/templates"
	"dumbbell/internal/utils"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func (s *HttpServer) startWorkoutHandler(w http.ResponseWriter, r *http.Request) {
	userId := s.SessionService.MustGetUserId(w, r)
	splitId, _ := strconv.ParseInt(r.FormValue("split"), 10, 64)

	workout, _ := dto.NewWorkout(splitId, userId, s.DB)

	pickExerciseData, err := s.WorkoutService.GetPickExerciseModel(userId, workout.ID)
	if err != nil {
		log.Printf("Error: in here %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Add("HX-Reselect", "#container")
	w.Header().Add("HX-Retarget", "#container")
	w.Header().Add("HX-Reswap", "outerHTML")
	w.Header().Add("HX-Trigger-After-Swap", "updateCharts")

	err = templates.ExecutePageTemplate(w, "pickExercise.html", pickExerciseData)
	if err != nil {
		log.Printf("Error: in here %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (s *HttpServer) startExerciseHandler(w http.ResponseWriter, r *http.Request) {
	userId := s.SessionService.MustGetUserId(w, r)
	workoutId := utils.MustParseInt64(r.FormValue("workoutId"))
	exerciseId := utils.MustParseInt64(r.FormValue("exercise"))

	exercise, err := dto.GetExercise(exerciseId, s.DB)
	if err != nil {
		log.Printf("Error getting exercise: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	workout, err := dto.GetWorkout(userId, workoutId, s.DB)
	if err != nil {
		log.Printf("Error getting workout: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = dto.CreateNewSet(workout.ID, exerciseId, s.DB)
	if err != nil {
		log.Printf("Error creating new set: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	sets := []dto.SetStatus{
		dto.SetCurrent,
	}
	for i := 1; i < int(exercise.Sets); i++ {
		sets = append(sets, dto.SetUncompleted)
	}

	viewModel := map[string]interface{}{
		"Title": fmt.Sprintf("Dumbbell - %s", exercise.Name),
		"Exercise": model.ExerciseViewModel{
			Name:        exercise.Name,
			WorkoutID:   workout.ID,
			Description: exercise.Description,
			ImageSrc:    exercise.GetImageURL(),
			WeightFrom:  exercise.WeightFrom,
			WeightTo:    exercise.WeightTo,
			RepsFrom:    exercise.RepsFrom,
			RepsTo:      exercise.RepsTo,
			Sets: model.ExerciseSetsModel{
				Items: sets,
				Htmx:  false,
			},
		},
	}
	templates.StartWorkout.Execute(w, viewModel)
}

func (s *HttpServer) nextExerciseHandler(w http.ResponseWriter, r *http.Request) {
	userId := s.SessionService.MustGetUserId(w, r)
	workoutId := utils.MustParseInt64(r.FormValue("workoutId"))
	rating := r.FormValue("rating")

	workout, getWorkoutErr := dto.GetWorkout(userId, workoutId, s.DB)
	if getWorkoutErr != nil {
		log.Printf("Error getting workout: %s", getWorkoutErr.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	newSet, createNextSetErr := dto.CreateNextSet(workout.ID, dto.SetStatus(rating), s.DB)
	if createNextSetErr != nil {
		if createNextSetErr == dto.ErrorSetLimitReached {
			pickExerciseData, pickExerciseModelErr := s.WorkoutService.GetPickExerciseModel(userId, workout.ID)
			if pickExerciseModelErr == nil {
				pickExerciseData.Header = s.SessionService.GetHeaderModel(r)
				templates.PickWorkout.Execute(w, pickExerciseData)
				return
			}

			if pickExerciseModelErr == service.ErrorNoExercises {
				if completeWorkoutErr := dto.CompleteWorkout(workout.ID, s.DB); completeWorkoutErr != nil {
					log.Printf("Error completing workout: %s", completeWorkoutErr.Error())
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				http.Redirect(w, r, "/", http.StatusMovedPermanently)
				return
			}
			log.Printf("Error getting pick exercise model: %s", pickExerciseModelErr.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Printf("Error creating next set: %s", createNextSetErr.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	completedSets, err := dto.GetCompletedWorkoutSets(workout.ID, newSet.ExerciseID, s.DB)
	if err != nil {
		log.Fatal(err.Error())
	}

	sets := []dto.SetStatus{}
	for _, completedSet := range completedSets {
		sets = append(sets, completedSet.SetRating)
	}
	sets = append(sets, newSet.SetRating)

	remaining := int(newSet.Sets) - len(sets)
	for i := 0; i < remaining; i++ {
		sets = append(sets, dto.SetUncompleted)
	}

	templates.NextExercise.Execute(w, model.ExerciseSetsModel{
		Items: sets,
		Htmx:  true,
	})
}

func (s *HttpServer) abortWorkout(w http.ResponseWriter, r *http.Request) {
	userId := s.SessionService.MustGetUserId(w, r)
	activeWorkout, err := dto.GetActiveWorkout(userId, s.DB)
	if err != nil {
		log.Printf("Error getting active workout: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = dto.DeleteWorkout(userId, activeWorkout.ID, s.DB)
	if err != nil {
		log.Printf("Error getting active workout: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func (s *HttpServer) workoutPageHandler(w http.ResponseWriter, r *http.Request) {
	userId := s.SessionService.MustGetUserId(w, r)
	activeWorkout, activeWorkoutErr := dto.GetActiveWorkout(userId, s.DB)
	if activeWorkoutErr == nil {
		activeWorkoutSet, activeWorkoutSetErr := dto.GetActiveWorkoutSet(activeWorkout.ID, s.DB)
		if activeWorkoutSetErr == nil {
			exercise, getExerciseErr := dto.GetExercise(activeWorkoutSet.ExerciseID, s.DB)
			if getExerciseErr != nil {
				log.Printf("Error getting exercise: %s", getExerciseErr.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			completedSets, getCompletedWorkoutsErr := dto.GetCompletedWorkoutSets(activeWorkoutSet.WorkoutID, activeWorkoutSet.ExerciseID, s.DB)
			if getCompletedWorkoutsErr != nil {
				log.Printf("Error getting completed workouts: %s", getCompletedWorkoutsErr.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			sets := []dto.SetStatus{}
			for _, completedSet := range completedSets {
				sets = append(sets, completedSet.SetRating)
			}
			sets = append(sets, activeWorkoutSet.SetRating)

			remaining := int(activeWorkoutSet.Sets) - len(sets)
			for i := 0; i < remaining; i++ {
				sets = append(sets, dto.SetUncompleted)
			}

			templateErr := templates.ExecutePageTemplate(w, "exercise.html", map[string]interface{}{
				"Title": fmt.Sprintf("Dumbbell - %s", exercise.Name),
				"Exercise": model.ExerciseViewModel{
					Name:        exercise.Name,
					WorkoutID:   activeWorkout.ID,
					Description: exercise.Description,
					ImageSrc:    exercise.GetImageURL(),
					WeightFrom:  exercise.WeightFrom,
					WeightTo:    exercise.WeightTo,
					RepsFrom:    exercise.RepsFrom,
					RepsTo:      exercise.RepsTo,
					Sets: model.ExerciseSetsModel{
						Items: sets,
						Htmx:  false,
					},
				},
			})
			if templateErr != nil {
				log.Printf("Error: in here %s", templateErr.Error())
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		pickExerciseData, pickExerciseErr := s.WorkoutService.GetPickExerciseModel(userId, activeWorkout.ID)
		if pickExerciseErr != nil {
			if pickExerciseErr == service.ErrorNoExercises {
				dto.CompleteWorkout(activeWorkout.ID, s.DB)
				w.Header().Add("HX-Replace-Url", "/")
				http.Redirect(w, r, "/", http.StatusMovedPermanently)
				return
			}

			log.Printf("Error: in here %s", pickExerciseErr.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		pickExerciseData.Header = s.SessionService.GetHeaderModel(r)
		templateErr := templates.ExecutePageTemplate(w, "pickExercise.html", pickExerciseData)
		if templateErr != nil {
			log.Printf("Error: in here %s", templateErr.Error())
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	log.Printf("Error getting active workout: %s", activeWorkoutErr.Error())
	w.Header().Add("HX-Replace-Url", "/")
	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}
