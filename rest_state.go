// rest_state.go

package main

import (
	"database/sql"
	"encoding/json"
	"github.com/Bnei-Baruch/gxydb-api/pkg/httputil"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"net/http"
	"strconv"
)

func (a *App) getProgram(w http.ResponseWriter, r *http.Request) {
	var s state
	vars := mux.Vars(r)
	s.Tag = "galaxy"
	s.StateID = "qids"
	id := vars["id"]

	if err := s.getProgram(a.DB, id); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, "Not Found")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondWithJSON(w, http.StatusOK, s.Data)
}

func (a *App) getGroups(w http.ResponseWriter, r *http.Request) {
	files, err := getGroups(a.DB)
	if err != nil {
		httputil.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.RespondWithJSON(w, http.StatusOK, files)
}

func (a *App) postGroup(w http.ResponseWriter, r *http.Request) {
	var i rooms
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		httputil.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	i.Room = id

	d := json.NewDecoder(r.Body)
	if err := d.Decode(&i); err != nil {
		httputil.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	defer r.Body.Close()

	if err := i.postGroup(a.DB); err != nil {
		httputil.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httputil.RespondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

func (a *App) findState(w http.ResponseWriter, r *http.Request) {
	key := r.FormValue("key")
	value := r.FormValue("value")

	states, err := findStates(a.DB, key, value)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, states)
}

func (a *App) getStateByTag(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tag := vars["tag"]

	states, err := getStateByTag(a.DB, tag)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, states)
}

func (a *App) getRooms(w http.ResponseWriter, r *http.Request) {

	states, err := getRooms(a.DB)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, states)
}

func (a *App) getRoom(w http.ResponseWriter, r *http.Request) {
	var i room
	vars := mux.Vars(r)
	id := vars["id"]

	err := i.getRoom(a.DB, id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, i)
}

func (a *App) getStates(w http.ResponseWriter, r *http.Request) {

	states, err := getStates(a.DB)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, states)
}

func (a *App) getState(w http.ResponseWriter, r *http.Request) {
	var s state
	vars := mux.Vars(r)
	s.StateID = vars["id"]

	if err := s.getState(a.DB); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, "Not Found")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondWithJSON(w, http.StatusOK, s.Data)
}

func (a *App) getStateJSON(w http.ResponseWriter, r *http.Request) {
	var s state
	vars := mux.Vars(r)
	s.Tag = vars["tag"]
	s.StateID = vars["id"]
	key := vars["jsonb"]

	if err := s.getStateJSON(a.DB, key); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, "Not Found")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondWithJSON(w, http.StatusOK, s.Data)
}

func (a *App) postState(w http.ResponseWriter, r *http.Request) {
	var s state
	vars := mux.Vars(r)
	s.Tag = vars["tag"]
	s.StateID = vars["id"]

	d := json.NewDecoder(r.Body)
	if err := d.Decode(&s.Data); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid resquest payload")
		return
	}

	defer r.Body.Close()

	if err := s.postState(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

func (a *App) updateState(w http.ResponseWriter, r *http.Request) {
	var s state
	vars := mux.Vars(r)
	s.StateID = vars["id"]

	d := json.NewDecoder(r.Body)
	if err := d.Decode(&s.Data); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid resquest payload")
		return
	}

	defer r.Body.Close()

	if err := s.updateState(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

func (a *App) postStateValue(w http.ResponseWriter, r *http.Request) {
	var s state
	vars := mux.Vars(r)
	s.StateID = vars["id"]
	key := vars["jsonb"]
	value := r.FormValue("value")
	status := r.FormValue("status")

	if value == "" {
		if err := s.postStateValue(a.DB, status, key); err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
	} else {
		if err := s.postStateStatus(a.DB, value, key); err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

func (a *App) postStateJSON(w http.ResponseWriter, r *http.Request) {
	var s state
	vars := mux.Vars(r)
	s.StateID = vars["id"]
	key := vars["jsonb"]
	var value map[string]interface{}
	d := json.NewDecoder(r.Body)

	if err := d.Decode(&value); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid resquest payload")
		return
	}

	if err := s.postStateJSON(a.DB, value, key); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

func (a *App) deleteState(w http.ResponseWriter, r *http.Request) {
	var s state
	vars := mux.Vars(r)
	s.StateID = vars["id"]

	if err := s.deleteState(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

func (a *App) deleteStateJSON(w http.ResponseWriter, r *http.Request) {
	var s state
	vars := mux.Vars(r)
	s.StateID = vars["id"]
	value := vars["jsonb"]

	if err := s.deleteStateJSON(a.DB, value); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}
