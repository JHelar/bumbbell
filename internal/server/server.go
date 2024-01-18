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

	newSet, err := dto.NextSet(workout.ID, dto.SetStatus(rating), s.DB)
	if err != nil {
		log.Fatal(err.Error())
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

	templates.ExecuteHtmxTemplate(w, "singleExerciseSets.html", sets)
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

	_, err = dto.NewSet(workout.ID, exerciseId, s.DB)
	if err != nil {
		log.Fatal(err.Error())
	}

	sets := []dto.SetStatus{
		dto.SetCurrent,
	}
	for i := 1; i <= int(exercise.Sets); i++ {
		sets = append(sets, dto.SetUncompleted)
	}

	templates.ExecuteHtmxTemplate(w, "singleExercise.html", model.ExerciseViewModel{
		Name:        exercise.Name,
		Description: exercise.Description,
		ImageSrc:    "/public/images/millitary_press.jpeg",
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

	// Create a new workout
	workout, _ := dto.NewWorkout(splitId, TEST_USER_ID, s.DB)
	log.Print(workout)

	cards := getExerciseCards(splitId, workout.ID, s.DB)
	templates.ExecuteHtmxTemplate(w, "pickExercise.html", cards)
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

func (s *HttpServer) homeHandler(w http.ResponseWriter, r *http.Request) {
	activeWorkout, err := dto.GetActiveWorkout(TEST_USER_ID, s.DB)
	if err == nil {
		activeWorkoutSet, err := dto.GetActiveWorkoutSet(activeWorkout.ID, s.DB)
		if err == nil {
			exercise, _ := dto.GetExercise(activeWorkoutSet.ExerciseID, s.DB)

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
					ImageSrc:    "/public/images/millitary_press.jpeg",
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
		templates.ExecutePageTemplate(w, "index.html", map[string]interface{}{
			"Title":     "Dumbell",
			"Exercises": cards,
		})
		return
	}

	cards := getSplitCards(s.DB)
	templates.ExecutePageTemplate(w, "index.html", map[string]interface{}{
		"Title":  "Dumbell",
		"Splits": cards,
	})

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
	handler.HandleFunc("/htmx/next", server.nextExerciseHandler)
	handler.HandleFunc("/htmx/exercise", server.startExerciseHandler)
	handler.HandleFunc("/htmx/split", server.startSplitHandler)
	handler.HandleFunc("/ws/hotreload", makeHMREndpoint())

	return &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}, nil
}
