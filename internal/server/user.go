package server

import (
	"dumbbell/internal/dto"
	"dumbbell/internal/model"
	"dumbbell/internal/templates"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func mustParseInt64(valueString string) int64 {
	value, err := strconv.ParseInt(valueString, 10, 64)
	if err != nil {
		log.Fatalf("mustParseInt64 Error: %s", err.Error())
	}
	return value
}

func mustParseFloat64(valueString string) float64 {
	value, err := strconv.ParseFloat(valueString, 64)
	if err != nil {
		log.Fatalf("mustParseFloat64 Error: %s", err.Error())
	}
	return value
}

func (s *HttpServer) newExercise(w http.ResponseWriter, r *http.Request) {
	splitId := mustParseInt64(r.FormValue("splitId"))
	templates.ExecuteHtmxTemplate(w, "editExercise.html", model.EditExerciseModel{
		SplitID: splitId,
	})
}

func (s *HttpServer) deleteExercise(w http.ResponseWriter, r *http.Request) {
	id := mustParseInt64(r.FormValue("id"))
	if err := dto.DeleteExercise(id, s.DB); err != nil {
		log.Printf("deleteExercise delete error: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *HttpServer) saveExercise(w http.ResponseWriter, r *http.Request) {
	var err error
	if err = r.ParseForm(); err != nil {
		log.Printf("saveExercise Parse form error: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	id := mustParseInt64(r.FormValue("id"))
	splitId := mustParseInt64(r.FormValue("splitId"))

	name := r.FormValue("name")
	description := r.FormValue("description")

	weightFrom := mustParseFloat64(r.FormValue("weight-from"))
	weightTo := mustParseFloat64(r.FormValue("weight-to"))
	repsFrom := mustParseInt64(r.FormValue("reps-from"))
	repsTo := mustParseInt64(r.FormValue("reps-to"))
	sets := mustParseInt64(r.FormValue("sets"))

	if err != nil {
		log.Printf("Error reading image file content: %s", err.Error())
	}

	imageReader, imageHeader, err := r.FormFile("image")
	var imageId *int64
	if err == nil {
		imageContentType := imageHeader.Header.Get("Content-Type")
		imageType := strings.Replace(imageContentType, "image/", "", 1)
		imageContent, err := io.ReadAll(imageReader)
		if err != nil {
			log.Printf("saveExercise error failed to read image data: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		image, err := dto.CreateImage(dto.ImageType(imageType), imageContent, s.DB)
		if err != nil {
			log.Printf("saveExercise error failed to create image: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		imageId = &image.ID
	} else {
		imageId = nil
	}

	isNew := id == 0

	var exercise dto.Exercise
	if isNew {
		exercise, err = dto.CreateExercise(splitId, imageId, name, description, weightFrom, weightTo, repsFrom, repsTo, sets, s.DB)
	} else {
		exercise, err = dto.UpdateExercise(id, name, imageId, description, weightFrom, weightTo, repsFrom, repsTo, sets, s.DB)
	}

	if err != nil {
		log.Printf("saveExercise error updating exercise: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}

	templates.ExecuteHtmxTemplate(w, "saveExercise.html", model.EditExerciseTableRowModel{
		ID:          exercise.ID,
		SplitID:     exercise.SplitID,
		Name:        exercise.Name,
		Description: exercise.Description,
		WeightFrom:  exercise.WeightFrom,
		WeightTo:    exercise.WeightTo,
		RepsFrom:    exercise.RepsFrom,
		RepsTo:      exercise.RepsTo,
		Sets:        exercise.Sets,
		ImageSrc:    exercise.GetImageURL(),
		IsNew:       isNew,
	})
}

func (s *HttpServer) editExercise(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Printf("editExercise Parse form error: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	exerciseId := mustParseInt64(r.FormValue("id"))

	exercise, err := dto.GetExercise(exerciseId, s.DB)
	if err != nil {
		log.Printf("editExercise GetExercise error: %s", err.Error())
	}

	templates.ExecuteHtmxTemplate(w, "editExercise.html", model.EditExerciseModel{
		ID:          exercise.ID,
		SplitID:     exercise.SplitID,
		Name:        exercise.Name,
		Description: exercise.Description,
		WeightFrom:  exercise.WeightFrom,
		WeightTo:    exercise.WeightTo,
		RepsFrom:    exercise.RepsFrom,
		RepsTo:      exercise.RepsTo,
		ImageSrc:    exercise.GetImageURL(),
		Sets:        exercise.Sets,
	})
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

		exerciseModels := []model.EditExerciseTableRowModel{}
		for _, exercise := range exercises {
			exerciseModels = append(exerciseModels, model.EditExerciseTableRowModel{
				ID:          exercise.ID,
				SplitID:     exercise.SplitID,
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

	if err = templates.ExecutePageTemplate(w, "user.html", model.UserSettingsModel{
		Title:  "Dumbbell - Settings",
		Splits: splitModels,
	}); err != nil {
		log.Printf("Error userHandler %s", err.Error())
	}
}
