package server

import (
	"database/sql"
	"dumbbell/internal/db"
	"dumbbell/internal/dto"
	"dumbbell/internal/model"
	"dumbbell/internal/templates"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type HttpServer struct {
	DB *sql.DB
}

var upgrader = websocket.Upgrader{}

const TEST_USER_ID int64 = 1

func reader(conn *websocket.Conn) {
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
	}
}

func makeHMREndpoint() func(http.ResponseWriter, *http.Request) {
	id := []byte(uuid.New().String())

	return func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
		}

		log.Println("HMR Client Connected")
		ws.WriteMessage(1, id)
		reader(ws)
	}
}

func (s *HttpServer) nextExerciseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Fatal(err.Error())
	}

	rating := r.Form.Get("rating")

	workout, err := dto.GetActiveWorkout(TEST_USER_ID, s.DB)
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

func (s *HttpServer) startExerciseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Fatal(err.Error())
	}

	exerciseIdString := r.Form.Get("exercise")
	exerciseId, _ := strconv.ParseInt(exerciseIdString, 10, 64)

	exercise, err := dto.GetExercise(exerciseId, s.DB)
	if err != nil {
		log.Fatal(err.Error())
	}

	workout, err := dto.GetActiveWorkout(TEST_USER_ID, s.DB)
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
		Description: exercise.Description,
		ImageSrc:    exercise.GetImageURL(),
		WeightFrom:  exercise.WeightFrom,
		WeightTo:    exercise.WeightTo,
		RepsFrom:    exercise.RepsFrom,
		RepsTo:      exercise.RepsTo,
		Sets:        sets,
	})
}

func getExerciseCards(splitId int64, workoutId int64, db *sql.DB) []model.CardViewModel {
	exercises, err := dto.GetAvailableExercises(splitId, workoutId, db)
	if err != nil {
		log.Fatal(err)
	}

	cards := []model.CardViewModel{}
	for _, exercise := range exercises {
		cards = append(cards, model.CardViewModel{
			ID:          fmt.Sprint(exercise.ID),
			Name:        exercise.Name,
			Description: exercise.Description,
		})
	}

	return cards
}

func (s *HttpServer) startSplitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Fatal(err.Error())
	}

	splitIdString := r.Form.Get("split")
	splitId, _ := strconv.ParseInt(splitIdString, 10, 64)

	workout, _ := dto.NewWorkout(splitId, TEST_USER_ID, s.DB)
	log.Print(workout)

	cards := getExerciseCards(splitId, workout.ID, s.DB)
	templates.ExecuteHtmxTemplate(w, "pickExercise.html", map[string]interface{}{
		"PageTitle": "Pick an exercise",
		"Cards":     cards,
	})
}

func getSplitCards(db *sql.DB) []model.CardViewModel {
	splits, err := dto.GetSplits(TEST_USER_ID, db)
	if err != nil {
		log.Fatal(err)
	}

	cards := []model.CardViewModel{}
	for _, split := range splits {
		cards = append(cards, model.CardViewModel{
			ID:          fmt.Sprint(split.ID),
			Name:        split.Name,
			Description: split.Description,
		})
	}

	return cards
}

func (s *HttpServer) handleExerciseImage(w http.ResponseWriter, r *http.Request) {
	urlParts := strings.Split(r.URL.Path, "/")
	exerciseIdString := urlParts[len(urlParts)-1]
	exerciseId, err := strconv.ParseInt(exerciseIdString, 10, 64)

	if err != nil {
		log.Print(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	image, err := dto.GetExerciseImage(exerciseId, s.DB)
	if err != nil {
		log.Print(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", fmt.Sprintf("image/%s", image.ContentType))
	w.Write(image.Content)
}

func NewServer() (*http.Server, error) {
	db, err := db.NewDB()
	if err != nil {
		return nil, err
	}
	server := &HttpServer{
		DB: db,
	}

	handler := http.NewServeMux()

	fs := http.FileServer(http.Dir("./public"))
	handler.Handle("/public/", http.StripPrefix("/public/", fs))

	handler.HandleFunc("/", server.homeHandler)
	handler.HandleFunc("/user", server.userHandler)
	handler.HandleFunc("/exercise/image/", server.handleExerciseImage)
	handler.HandleFunc("/htmx/next", server.nextExerciseHandler)
	handler.HandleFunc("/htmx/exercise", server.startExerciseHandler)
	handler.HandleFunc("/htmx/exercise/edit", server.editExercise)
	handler.HandleFunc("/htmx/exercise/edit/save", server.saveExercise)
	handler.HandleFunc("/htmx/split", server.startSplitHandler)
	handler.HandleFunc("/ws/hotreload", makeHMREndpoint())

	return &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}, nil
}
