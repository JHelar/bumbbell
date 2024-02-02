package server

import (
	"dumbbell/internal/dto"
	"dumbbell/internal/model"
	"dumbbell/internal/service"
	"dumbbell/internal/templates"
	"dumbbell/internal/utils"
	"log"
	"net/http"
	"strconv"
)

func (s *HttpServer) startWorkoutHandler(w http.ResponseWriter, r *http.Request) {
	splitId, _ := strconv.ParseInt(r.FormValue("split"), 10, 64)

	workout, _ := dto.NewWorkout(splitId, TEST_USER_ID, s.DB)

	pickExerciseData, err := s.WorkoutService.GetPickExerciseModel(TEST_USER_ID, workout.ID)
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
	workoutId := utils.MustParseInt64(r.FormValue("workoutId"))
	exerciseId := utils.MustParseInt64(r.FormValue("exercise"))

	exercise, err := dto.GetExercise(exerciseId, s.DB)
	if err != nil {
		log.Fatal(err.Error())
	}

	workout, err := dto.GetWorkout(TEST_USER_ID, workoutId, s.DB)
	if err != nil {
		log.Fatal(err.Error())
	}

	_, err = dto.CreateNewSet(workout.ID, exerciseId, s.DB)
	if err != nil {
		log.Fatal(err.Error())
	}

	sets := []dto.SetStatus{
		dto.SetCurrent,
	}
	for i := 1; i < int(exercise.Sets); i++ {
		sets = append(sets, dto.SetUncompleted)
	}

	w.Header().Add("HX-Reselect", "#container")
	w.Header().Add("HX-Retarget", "#container")
	w.Header().Add("HX-Reswap", "outerHTML")
	templates.ExecutePageTemplate(w, "exercise.html", map[string]interface{}{
		"Title": "Dumbell",
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
	})
}

func (s *HttpServer) nextExerciseHandler(w http.ResponseWriter, r *http.Request) {
	workoutId := utils.MustParseInt64(r.FormValue("workoutId"))
	rating := r.FormValue("rating")

	workout, err := dto.GetWorkout(TEST_USER_ID, workoutId, s.DB)
	if err != nil {
		log.Fatal(err.Error())
	}

	newSet, err := dto.CreateNextSet(workout.ID, dto.SetStatus(rating), s.DB)
	if err != nil {
		if err == dto.ErrorSetLimitReached {
			pickExerciseData, err := s.WorkoutService.GetPickExerciseModel(TEST_USER_ID, workout.ID)
			if err == nil {
				w.Header().Add("HX-Reselect", "#container")
				w.Header().Add("HX-Retarget", "#container")
				w.Header().Add("HX-Reswap", "outerHTML")
				w.Header().Add("HX-Trigger-After-Swap", "updateCharts")

				templates.ExecutePageTemplate(w, "pickExercise.html", pickExerciseData)
				return
			}

			if err == service.ErrorNoExercises {
				if err = dto.CompleteWorkout(workout.ID, s.DB); err != nil {
					log.Print("Unable to complete workout!")
					return
				}

				cards := getSplitCards(s.DB)
				w.Header().Add("HX-Reselect", "#container")
				w.Header().Add("HX-Retarget", "#container")
				w.Header().Add("HX-Reswap", "outerHTML")
				w.Header().Add("HX-Trigger-After-Swap", "updateCharts")

				templates.ExecutePageTemplate(w, "workout.html", map[string]interface{}{
					"Splits": cards,
				})
				return
			}
		}
		log.Printf(err.Error())
		http.Redirect(w, r, "/", http.StatusTeapot)
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

	templates.ExecuteHtmxTemplate(w, "nextExercise.html", model.ExerciseSetsModel{
		Items: sets,
		Htmx:  true,
	})
}

func (s *HttpServer) abortWorkout(w http.ResponseWriter, r *http.Request) {
	activeWorkout, err := dto.GetActiveWorkout(TEST_USER_ID, s.DB)
	if err != nil {
		log.Printf("Error getting active workout: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = dto.DeleteWorkout(TEST_USER_ID, activeWorkout.ID, s.DB)
	if err != nil {
		log.Printf("Error getting active workout: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func (s *HttpServer) workoutPageHandler(w http.ResponseWriter, r *http.Request) {
	activeWorkout, err := dto.GetActiveWorkout(TEST_USER_ID, s.DB)
	if err == nil {
		activeWorkoutSet, err := dto.GetActiveWorkoutSet(activeWorkout.ID, s.DB)
		if err == nil {
			exercise, err := dto.GetExercise(activeWorkoutSet.ExerciseID, s.DB)
			if err != nil {
				log.Print("Could not get associated exercise")
			}

			completedSets, err := dto.GetCompletedWorkoutSets(activeWorkoutSet.WorkoutID, activeWorkoutSet.ExerciseID, s.DB)
			if err != nil {
				log.Fatal(err.Error())
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

			err = templates.ExecutePageTemplate(w, "exercise.html", map[string]interface{}{
				"Title": "Dumbell",
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
			return
		}

		pickExerciseData, err := s.WorkoutService.GetPickExerciseModel(TEST_USER_ID, activeWorkout.ID)
		if err != nil {
			if err == service.ErrorNoExercises {
				dto.CompleteWorkout(activeWorkout.ID, s.DB)
				http.Redirect(w, r, "/", http.StatusMovedPermanently)
				return
			}

			log.Printf("Error: in here %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = templates.ExecutePageTemplate(w, "pickExercise.html", pickExerciseData)
		if err != nil {
			log.Printf("Error: in here %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	} else {
		log.Print("No active workout")
	}

	cards := getSplitCards(s.DB)
	templates.ExecutePageTemplate(w, "workout.html", map[string]interface{}{
		"Title":     "Dumbell",
		"PageTitle": "Start a workout",
		"Splits":    cards,
	})
}
