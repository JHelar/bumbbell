package server

import (
	"dumbbell/internal/dto"
	"dumbbell/internal/model"
	"dumbbell/internal/templates"
	"log"
	"net/http"
)

func (s *HttpServer) homeHandler(w http.ResponseWriter, r *http.Request) {
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

			templates.ExecutePageTemplate(w, "index.html", map[string]interface{}{
				"Title": "Dumbell",
				"Exercise": model.ExerciseViewModel{
					Name:        exercise.Name,
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
			templates.ExecutePageTemplate(w, "index.html", map[string]interface{}{
				"Title":     "Dumbell",
				"PageTitle": "Pick an exercise",
				"Exercises": cards,
			})
			return
		}
	} else {
		log.Print("No active workout")
	}

	cards := getSplitCards(s.DB)
	templates.ExecutePageTemplate(w, "index.html", map[string]interface{}{
		"Title":     "Dumbell",
		"PageTitle": "Start a workout",
		"Splits":    cards,
	})
}
