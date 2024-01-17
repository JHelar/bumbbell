package server

import (
	"dumbbell/internal/db"
	"dumbbell/internal/dto"
	"dumbbell/internal/model"
	"dumbbell/internal/templates"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type HttpServer struct {
	DB *db.DB
}

var upgrader = websocket.Upgrader{}

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

func nextExerciseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Fatal(err.Error())
	}

	rating := r.Form.Get("rating")
	templates.ExecuteHtmxTemplate(w, "singleExercise.html", model.ExerciseViewModel{
		Name:        "Military press",
		Description: "Barbell, squat rack",
		ImageSrc:    "/public/images/millitary_press.jpeg",
		WeightFrom:  0,
		WeightTo:    7.25,
		RepsFrom:    12,
		RepsTo:      20,
		Sets: []dto.SetStatus{
			dto.SetStatus(rating),
			dto.SetCurrent,
			dto.SetUncompleted,
		},
	})
}

func startExerciseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Fatal(err.Error())
	}

	exerciseName := r.Form.Get("exercise")
	log.Println(exerciseName)

	templates.ExecuteHtmxTemplate(w, "singleExercise.html", model.ExerciseViewModel{
		Name:        "Military press",
		Description: "Barbell, squat rack",
		ImageSrc:    "/public/images/millitary_press.jpeg",
		WeightFrom:  0,
		WeightTo:    7.25,
		RepsFrom:    12,
		RepsTo:      20,
		Sets: []dto.SetStatus{
			dto.SetCurrent,
			dto.SetUncompleted,
			dto.SetUncompleted,
		},
	})
}

func startSplitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err := r.ParseForm(); err != nil {
		log.Fatal(err.Error())
	}

	splitName := r.Form.Get("split")
	log.Println(splitName)

	templates.ExecuteHtmxTemplate(w, "pickExercise.html", []model.CardViewModel{
		{
			Name:        "Millitary press",
			Description: "Barbell, squat rack",
		},
	})
}

func (s *HttpServer) homeHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecutePageTemplate(w, "index.html", map[string]interface{}{
		"Title": "Dumbell",
		"Splits": []model.CardViewModel{
			{
				Name:        "Shoulders",
				Description: "Shoulder workout",
			},
		},
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
	handler.HandleFunc("/htmx/next", nextExerciseHandler)
	handler.HandleFunc("/htmx/start", startExerciseHandler)
	handler.HandleFunc("/htmx/split", startSplitHandler)
	handler.HandleFunc("/ws/hotreload", makeHMREndpoint())

	return &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}, nil
}
