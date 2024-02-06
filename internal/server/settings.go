package server

import (
	"dumbbell/internal/dto"
	"dumbbell/internal/model"
	"dumbbell/internal/templates"
	"dumbbell/internal/utils"
	"io"
	"log"
	"net/http"
	"strings"
)

func (s *HttpServer) newSplit(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("HX-Reswap", "beforeend")
	w.Header().Add("HX-Retarget", "main")
	templates.ExecuteHtmxTemplate(w, "editSplit.html", model.EditSplitModel{})
}

func (s *HttpServer) editSplit(w http.ResponseWriter, r *http.Request) {
	userId := s.SessionService.MustGetUserId(w, r)
	splitId := utils.MustParseInt64(r.FormValue("splitId"))

	split, err := dto.GetSplit(userId, splitId, s.DB)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("HX-Reswap", "beforeend")
	w.Header().Add("HX-Retarget", "main")
	templates.ExecuteHtmxTemplate(w, "editSplit.html", model.EditSplitModel{
		ID:          split.ID,
		Name:        split.Name,
		Description: split.Description,
	})
}

func (s *HttpServer) saveSplit(w http.ResponseWriter, r *http.Request) {
	userId := s.SessionService.MustGetUserId(w, r)
	splitId := utils.MustParseInt64(r.FormValue("splitId"))

	name := r.FormValue("name")
	description := r.FormValue("description")

	if splitId == 0 {
		split, err := dto.CreateSplit(userId, name, description, s.DB)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = templates.ExecuteHtmxTemplate(w, "newSplit.html", model.EditWorkoutTableSplitModel{
			ID:          split.ID,
			Name:        split.Name,
			Description: split.Description,
			Exercises:   []model.EditExerciseTableRowModel{},
		})

		if err != nil {
			log.Printf("Error in save split template: %s", err.Error())
		}

	} else {
		split, err := dto.UpdateSplit(userId, splitId, name, description, s.DB)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		template, err := templates.New("SaveSplitResponse").Parse(`
		<h3 hx-swap-oob="innerHTML:#split-{{ .ID }} h3">{{ .Name }}</h3>
		<p hx-swap-oob="innerHTML:#split-{{ .ID }} h3+p">{{ .Description }}</p>
		`)

		if err != nil {
			log.Printf("Error in save split template: %s", err.Error())
		}
		template.Execute(w, struct {
			ID          int64
			Name        string
			Description string
		}{ID: split.ID, Name: split.Name, Description: split.Description})
	}
}

func (s *HttpServer) deleteSplit(w http.ResponseWriter, r *http.Request) {
	userId := s.SessionService.MustGetUserId(w, r)
	splitId := utils.MustParseInt64(r.FormValue("splitId"))

	if err := dto.DeleteSplit(userId, splitId, s.DB); err != nil {
		log.Printf("Delete split error: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *HttpServer) newExercise(w http.ResponseWriter, r *http.Request) {
	splitId := utils.MustParseInt64(r.FormValue("splitId"))

	w.Header().Add("HX-Reswap", "beforeend")
	w.Header().Add("HX-Retarget", "main")
	templates.ExecuteHtmxTemplate(w, "editExercise.html", model.EditExerciseModel{
		SplitID: splitId,
	})
}

func (s *HttpServer) deleteExercise(w http.ResponseWriter, r *http.Request) {
	id := utils.MustParseInt64(r.FormValue("id"))
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

	id := utils.MustParseInt64(r.FormValue("id"))
	splitId := utils.MustParseInt64(r.FormValue("splitId"))

	name := r.FormValue("name")
	description := r.FormValue("description")

	weightFrom := utils.MustParseFloat64(r.FormValue("weight-from"))
	weightTo := utils.MustParseFloat64(r.FormValue("weight-to"))
	repsFrom := utils.MustParseInt64(r.FormValue("reps-from"))
	repsTo := utils.MustParseInt64(r.FormValue("reps-to"))
	sets := utils.MustParseInt64(r.FormValue("sets"))

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

	exerciseId := utils.MustParseInt64(r.FormValue("id"))

	exercise, err := dto.GetExercise(exerciseId, s.DB)
	if err != nil {
		log.Printf("editExercise GetExercise error: %s", err.Error())
	}

	w.Header().Add("HX-Reswap", "beforeend")
	w.Header().Add("HX-Retarget", "main")
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

func (s *HttpServer) settingsPageHandler(w http.ResponseWriter, r *http.Request) {
	userId := s.SessionService.MustGetUserId(w, r)
	splits, err := dto.GetSplits(userId, s.DB)

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
		Header: s.SessionService.GetHeaderModel(r),
	}); err != nil {
		log.Printf("Error userHandler %s", err.Error())
	}
}
