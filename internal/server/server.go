package server

import (
	"database/sql"
	"dumbbell/internal/db"
	"dumbbell/internal/dto"
	"dumbbell/internal/model"
	"dumbbell/internal/mux"
	"dumbbell/internal/service"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type HttpServer struct {
	DB              *sql.DB
	WorkoutService  service.WorkoutService
	ExerciseService service.ExerciseService
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

func getExerciseCards(splitId int64, workoutId int64, db *sql.DB) []model.CardViewModel {
	exercises, err := dto.GetAvailableExercises(splitId, workoutId, db)
	if err != nil {
		log.Fatal(err)
	}

	cards := []model.CardViewModel{}
	for _, exercise := range exercises {
		cards = append(cards, model.CardViewModel{
			ID:          exercise.ID,
			WorkoutID:   workoutId,
			Name:        exercise.Name,
			Description: exercise.Description,
		})
	}

	return cards
}

func getSplitCards(db *sql.DB) []model.CardViewModel {
	splits, err := dto.GetSplits(TEST_USER_ID, db)
	if err != nil {
		log.Fatal(err)
	}

	cards := []model.CardViewModel{}
	for _, split := range splits {
		cards = append(cards, model.CardViewModel{
			ID:          split.ID,
			Name:        split.Name,
			Description: split.Description,
		})
	}

	return cards
}

func (s *HttpServer) handleExerciseImage(w http.ResponseWriter, r *http.Request) {
	exerciseIdString := r.FormValue("id")
	exerciseId, err := strconv.ParseInt(exerciseIdString, 10, 64)

	if err != nil {
		log.Printf("Error parse exercise id url: %s", r.URL.Path)
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
		DB:              db,
		WorkoutService:  service.NewWorkoutService(db),
		ExerciseService: service.NewExerciseService(db),
	}

	handler := mux.NewHttpMux()

	fs := http.FileServer(http.Dir("./public"))
	handler.Handle("/public/", http.StripPrefix("/public/", fs))

	handler.HandleFunc("/user", server.userHandler)
	handler.HandleFunc("/exercise/image/(?P<id>[\\d]+)", server.handleExerciseImage)

	handler.HandleFunc("/workout", server.workoutPageHandler)

	handler.PostFunc("/htmx/workout/start", server.startWorkoutHandler)
	handler.DeleteFunc("/htmx/workout/abort", server.abortWorkout)
	handler.PostFunc("/htmx/workout/(?P<workoutId>[\\d]+)/exercise/start", server.startExerciseHandler)
	handler.PostFunc("/htmx/workout/(?P<workoutId>[\\d]+)/exercise/next", server.nextExerciseHandler)

	handler.GetFunc("/htmx/split/new", server.newSplit)
	handler.GetFunc("/htmx/split/(?P<splitId>[\\d]+)/edit", server.editSplit)
	handler.PostFunc("/htmx/split/(?P<splitId>[\\d]+)/save", server.saveSplit)
	handler.DeleteFunc("/htmx/split/(?P<splitId>[\\d]+)/delete", server.deleteSplit)

	handler.GetFunc("/htmx/split/(?P<splitId>[\\d]+)/exercise/new", server.newExercise)
	handler.GetFunc("/htmx/split/(?P<splitId>[\\d]+)/exercise/(?P<id>[\\d]+)/edit", server.editExercise)
	handler.PostFunc("/htmx/split/(?P<splitId>[\\d]+)/exercise/(?P<id>[\\d]+)/save", server.saveExercise)
	handler.DeleteFunc("/htmx/split/(?P<splitId>[\\d]+)/exercise/(?P<id>[\\d]+)/delete", server.deleteExercise)

	handler.HandleFunc("/ws/hotreload", makeHMREndpoint())
	handler.HandleFunc("/", server.homeHandler)

	return &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}, nil
}
