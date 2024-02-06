package server

import (
	"database/sql"
	"dumbbell/internal/db"
	"dumbbell/internal/dto"
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
	SessionService  service.SessionService
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
		SessionService:  service.NewSessionService(db),
	}

	handler := mux.NewHttpMux("")

	fs := http.FileServer(http.Dir("./public"))
	handler.Handle("/public/", http.StripPrefix("/public/", fs))

	userRouter := handler.Use("/user", server.SessionService.AuthMiddleware)
	userRouter.HandleFunc("", server.settingsPageHandler)

	handler.HandleFunc("/exercise/image/(?P<id>[\\d]+)", server.handleExerciseImage)

	workoutRouter := handler.Use("/workout", server.SessionService.AuthMiddleware)
	workoutRouter.HandleFunc("", server.workoutPageHandler)
	workoutRouter.PostFunc("/start", server.startWorkoutHandler)
	workoutRouter.DeleteFunc("/abort", server.abortWorkout)
	workoutRouter.PostFunc("/(?P<workoutId>[\\d]+)/exercise/start", server.startExerciseHandler)
	workoutRouter.PostFunc("/(?P<workoutId>[\\d]+)/exercise/next", server.nextExerciseHandler)

	settingsRouter := handler.Use("/split", server.SessionService.AuthMiddleware)
	settingsRouter.GetFunc("/new", server.newSplit)
	settingsRouter.GetFunc("/(?P<splitId>[\\d]+)/edit", server.editSplit)
	settingsRouter.PostFunc("/(?P<splitId>[\\d]+)/save", server.saveSplit)
	settingsRouter.DeleteFunc("/(?P<splitId>[\\d]+)/delete", server.deleteSplit)

	settingsRouter.GetFunc("/(?P<splitId>[\\d]+)/exercise/new", server.newExercise)
	settingsRouter.GetFunc("/(?P<splitId>[\\d]+)/exercise/(?P<id>[\\d]+)/edit", server.editExercise)
	settingsRouter.PostFunc("/(?P<splitId>[\\d]+)/exercise/(?P<id>[\\d]+)/save", server.saveExercise)
	settingsRouter.DeleteFunc("/(?P<splitId>[\\d]+)/exercise/(?P<id>[\\d]+)/delete", server.deleteExercise)

	handler.GetFunc("/login", server.loginPageHandler)
	handler.PostFunc("/login", server.LoginUser)

	handler.GetFunc("/signup", server.signupPageHandler)
	handler.PostFunc("/signup", server.RegisterUser)
	handler.GetFunc("/logout", server.LogoutUser)

	handler.HandleFunc("/ws/hotreload", makeHMREndpoint())
	handler.HandleFunc("/", server.homeHandler)

	return &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}, nil
}
