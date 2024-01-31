package server

import (
	"dumbbell/internal/dto"
	"dumbbell/internal/model"
	"dumbbell/internal/templates"
	"dumbbell/internal/utils"
	"log"
	"net/http"
	"strconv"
)

func (s *HttpServer) startWorkoutHandler(w http.ResponseWriter, r *http.Request) {
	splitId, _ := strconv.ParseInt(r.FormValue("split"), 10, 64)

	workout, _ := dto.NewWorkout(splitId, TEST_USER_ID, s.DB)
	log.Print(workout)

	cards := getExerciseCards(splitId, workout.ID, s.DB)
	templates.ExecuteHtmxTemplate(w, "pickExercise.html", map[string]interface{}{
		"PageTitle": "Pick an exercise",
		"Cards":     cards,
		"WorkoutID": workout.ID,
	})
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

	templates.ExecuteHtmxTemplate(w, "startExercise.html", model.ExerciseViewModel{
		Name:        exercise.Name,
		WorkoutID:   workoutId,
		Description: exercise.Description,
		ImageSrc:    exercise.GetImageURL(),
		WeightFrom:  exercise.WeightFrom,
		WeightTo:    exercise.WeightTo,
		RepsFrom:    exercise.RepsFrom,
		RepsTo:      exercise.RepsTo,
		Sets:        sets,
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
			cards := getExerciseCards(workout.SplitID, workout.ID, s.DB)
			if len(cards) > 0 {
				templates.ExecuteHtmxTemplate(w, "pickExercise.html", map[string]interface{}{
					"PageTitle": "Pick an Exercise",
					"Cards":     cards,
				})
				return
			}

			if err = dto.CompleteWorkout(workout.ID, s.DB); err != nil {
				log.Print("Unable to complete workout!")
				return
			}
			cards = getSplitCards(s.DB)
			templates.ExecuteHtmxTemplate(w, "pickSplit.html", map[string]interface{}{
				"PageTitle": "Start a workout",
				"Cards":     cards,
			})
			return
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

	templates.ExecuteHtmxTemplate(w, "nextExercise.html", sets)
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

			templates.ExecutePageTemplate(w, "workout.html", map[string]interface{}{
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
					Sets:        sets,
				},
			})
			return
		}

		cards := getExerciseCards(activeWorkout.SplitID, activeWorkout.ID, s.DB)
		if len(cards) > 0 {
			err := templates.ExecutePageTemplate(w, "workout.html", map[string]interface{}{
				"Title":     "Dumbell",
				"PageTitle": "Pick an exercise",
				"Exercises": cards,
			})
			if err != nil {
				log.Printf("Error: in here %s", err.Error())
			}
			return
		} else {
			log.Print("Delete workout")
			dto.DeleteWorkout(TEST_USER_ID, activeWorkout.ID, s.DB)
		}
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
