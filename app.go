// app.go

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type App struct {
	Router *mux.Router
	DB     *sql.DB
}

func (a *App) Initialize(user string, password string, dbname string) {
	connectionString :=
		fmt.Sprintf("postgres://%s:%s@localhost/%s?sslmode=disable", user, password, dbname)

	var err error
	a.DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	a.Router = mux.NewRouter()
	a.initializeRoutes()
}

func (a *App) Run(addr string) {
	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Content-Length", "Accept-Encoding"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "DELETE", "POST", "PUT", "OPTIONS"})

	log.Fatal(http.ListenAndServe(":8880", handlers.CORS(originsOk, headersOk, methodsOk)(a.Router)))
}

func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/states", a.getStates).Methods("GET")
	a.Router.HandleFunc("/galaxy/program/{id}", a.getProgram).Methods("GET")
	a.Router.HandleFunc("/galaxy/groups", a.getGroups).Methods("GET")
	a.Router.HandleFunc("/galaxy/group/{id}", a.postGroup).Methods("PUT")
	a.Router.HandleFunc("/galaxy/rooms", a.getRooms).Methods("GET")
	a.Router.HandleFunc("/galaxy/room/{id}", a.getRoom).Methods("GET")
	a.Router.HandleFunc("/{tag}", a.getStateByTag).Methods("GET")
	a.Router.HandleFunc("/{tag}/{id}", a.getState).Methods("GET")
	a.Router.HandleFunc("/{tag}/{id}/{jsonb}", a.getStateJSON).Methods("GET")
	a.Router.HandleFunc("/{tag}/{id}", a.postState).Methods("PUT")
	a.Router.HandleFunc("/{tag}/{id}", a.updateState).Methods("POST")
	a.Router.HandleFunc("/{tag}/{id}/{jsonb}", a.postStateJSON).Methods("PUT")
	a.Router.HandleFunc("/{tag}/{id}/{jsonb}", a.postStateValue).Methods("POST")
	a.Router.HandleFunc("/{tag}/{id}", a.deleteState).Methods("DELETE")
	a.Router.HandleFunc("/{tag}/{id}/{jsonb}", a.deleteStateJSON).Methods("DELETE")
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(code)
	w.Write(response)
}
