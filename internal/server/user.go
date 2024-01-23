package server

import (
	"dumbbell/internal/dto"
	"dumbbell/internal/model"
	"dumbbell/internal/templates"
	"log"
	"net/http"
)

func (s *HttpServer) saveExercise(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteHtmxTemplate(w, "saveExercise.html", nil)
}

func (s *HttpServer) editExercise(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteHtmxTemplate(w, "editExercise.html", nil)
}

func (s *HttpServer) userHandler(w http.ResponseWriter, r *http.Request) {
	splits, err := dto.GetSplits(TEST_USER_ID, s.DB)

	if err != nil {
		log.Printf("Error userHandler %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	splitModels := []model.EditWorkoutTableSplitModel{}
	for _, split := range splits {
		exercises, err := dto.GetAllExercises(split.ID, s.DB)
		if err != nil {
			log.Printf("Error userHandler %s", err.Error())
			break
		}

		exerciseModels := []model.EditWorkoutTableExerciseModel{}
		for _, exercise := range exercises {
			exerciseModels = append(exerciseModels, model.EditWorkoutTableExerciseModel{
				ID:          exercise.ID,
				Name:        exercise.Name,
				Description: exercise.Description,
				WeightFrom:  exercise.WeightFrom,
				WeightTo:    exercise.WeightTo,
				RepsFrom:    exercise.RepsFrom,
				RepsTo:      exercise.RepsTo,
				Sets:        exercise.Sets,
				ImageSrc:    exercise.GetImageURL(),
			})
		}

		splitModels = append(splitModels, model.EditWorkoutTableSplitModel{
			ID:          split.ID,
			Name:        split.Name,
			Description: split.Description,
			Exercises:   exerciseModels,
		})
	}

	templates.ExecutePageTemplate(w, "user.html", model.UserSettingsModel{
		Title:  "Dumbbell - Settings",
		Splits: splitModels,
	})
}
