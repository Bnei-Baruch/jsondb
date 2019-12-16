// model_state.go

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
)

type state struct {
	ID      int                    `json:"id"`
	StateID string                 `json:"state_id"`
	Data    map[string]interface{} `json:"data"`
	Tag     string                 `json:"tag"`
}

type room struct {
	Room      int         `json:"room"`
	Group     string      `json:"description"`
	Questions bool        `json:"questions"`
	NumUsers  int         `json:"num_users"`
	Users     interface{} `json:"users"`
}

func getRooms(db *sql.DB) ([]room, error) {
	rows, err := db.Query(
		"with data as (SELECT jsonb_path_query(data, '$.*') as data FROM state WHERE state_id = 'users') select * from (select distinct on (room) (data -> 'room')::text::bigint as room,(data -> 'group')::text as group,(data -> 'timestamp')::text::bigint as stamp from data where (data -> 'room') is not null order by room,stamp)p order by stamp;")

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	rooms := []room{}

	for rows.Next() {
		var r room
		var st int
		var o interface{}
		var obj []byte
		var grp []byte
		if err := rows.Scan(&r.Room, &grp, &st); err != nil {
			return nil, err
		}

		uq := fmt.Sprintf("SELECT jsonb_path_query_array(data, '$.* ? (@.room == %v)') FROM state WHERE state_id = 'users'", r.Room)
		qq := fmt.Sprintf("SELECT jsonb_path_exists(data, '$.* ? (@.room == %v && @.question == true)') FROM state WHERE state_id = 'users'", r.Room)
		err := db.QueryRow(uq).Scan(&obj)
		if err != nil {
			return nil, err
		}
		err = db.QueryRow(qq).Scan(&r.Questions)
		if err != nil {
			return nil, err
		}
		json.Unmarshal(obj, &o)
		json.Unmarshal(grp, &r.Group)
		r.Users = o
		r.NumUsers = len(o.([]interface{}))
		rooms = append(rooms, r)
	}

	return rooms, nil
}

func (r *room) getRoom(db *sql.DB, id string) error {
	var o interface{}
	var obj []byte
	var grp []byte
	rid, _ := strconv.Atoi(id)

	rq := fmt.Sprintf("with data as (SELECT jsonb_path_query_first(data, '$.* ? (@.room == %v)') as data FROM state WHERE state_id = 'users') select * from (select  (data -> 'room')::text::bigint as room, (data -> 'group')::text as group from data) p", rid)
	uq := fmt.Sprintf("SELECT jsonb_path_query_array(data, '$.* ? (@.room == %v)') FROM state WHERE state_id = 'users'", rid)
	qq := fmt.Sprintf("SELECT jsonb_path_exists(data, '$.* ? (@.room == %v && @.question == true)') FROM state WHERE state_id = 'users'", rid)
	err := db.QueryRow(rq).Scan(&r.Room, &grp)
	if err != nil {
		return err
	}

	err = db.QueryRow(uq).Scan(&obj)
	if err != nil {
		return err
	}

	err = db.QueryRow(qq).Scan(&r.Questions)
	if err != nil {
		return err
	}

	json.Unmarshal(obj, &o)
	json.Unmarshal(grp, &r.Group)
	r.Users = o
	r.NumUsers = len(o.([]interface{}))

	return nil
}

func findStates(db *sql.DB, key string, value string) ([]state, error) {
	rows, err := db.Query(
		"SELECT id, state_id, data FROM state WHERE data @> json_build_object($1::text, $2::text)::jsonb",
		key, value)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	states := []state{}

	for rows.Next() {
		var s state
		var obj []byte
		if err := rows.Scan(&s.ID, &s.StateID, &obj); err != nil {
			return nil, err
		}
		json.Unmarshal(obj, &s.Data)
		states = append(states, s)
	}

	return states, nil
}

func getStates(db *sql.DB) ([]state, error) {
	rows, err := db.Query(
		"SELECT id, state_id, data, tag FROM state ORDER BY tag")

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	states := []state{}

	for rows.Next() {
		var s state
		var obj []byte
		if err := rows.Scan(&s.ID, &s.StateID, &obj, &s.Tag); err != nil {
			return nil, err
		}
		json.Unmarshal(obj, &s.Data)
		states = append(states, s)
	}

	return states, nil
}

func getStateByTag(db *sql.DB, tag string) (map[string]interface{}, error) {
	rows, err := db.Query(
		"SELECT id, state_id, data FROM state WHERE tag = $1 ORDER BY state_id DESC",
		tag)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	states := make(map[string]interface{})

	for rows.Next() {
		var s state
		var o map[string]interface{}
		var obj []byte
		if err := rows.Scan(&s.ID, &s.StateID, &obj); err != nil {
			return nil, err
		}
		json.Unmarshal(obj, &o)
		states[s.StateID] = o
	}

	return states, nil
}

func (s *state) getState(db *sql.DB) error {
	var obj []byte
	err := db.QueryRow("SELECT data FROM state WHERE state_id = $1",
		s.StateID).Scan(&obj)
	if err != nil {
		return err
	}
	err = json.Unmarshal(obj, &s.Data)

	return err
}

func (s *state) getStateJSON(db *sql.DB, key string) error {
	var obj []byte
	err := db.QueryRow("SELECT data->>$2 FROM state where state_id = $1",
		s.StateID, key).Scan(&obj)
	if err != nil {
		return err
	}
	err = json.Unmarshal(obj, &s.Data)

	return err
}

func (s *state) postState(db *sql.DB) error {
	v, _ := json.Marshal(s.Data)

	err := db.QueryRow(
		"INSERT INTO state(state_id, data, tag) VALUES($1, $2, $3) ON CONFLICT (state_id) DO UPDATE SET data = $2 WHERE state.state_id = $1 RETURNING id",
		s.StateID, v, s.Tag).Scan(&s.ID)

	if err != nil {
		return err
	}

	return nil
}

func (s *state) updateState(db *sql.DB) error {
	v, _ := json.Marshal(s.Data)
	_, err :=
		db.Exec("UPDATE state SET data=$2 WHERE state_id=$1",
			s.StateID, v)

	return err
}

func (s *state) postStateStatus(db *sql.DB, value, key string) error {
	_, err := db.Exec("UPDATE state SET data = data || json_build_object($3::text, $2::bool)::jsonb WHERE state_id=$1",
		s.StateID, value, key)

	return err
}

func (s *state) postStateValue(db *sql.DB, value string, key string) error {
	_, err := db.Exec("UPDATE state SET data = data || json_build_object($3::text, $2::text)::jsonb WHERE state_id=$1",
		s.StateID, value, key)

	return err
}

func (s *state) postStateJSON(db *sql.DB, value interface{}, key string) error {
	v, _ := json.Marshal(value)
	_, err := db.Exec("UPDATE state SET data = data || json_build_object($3::text, $2::jsonb)::jsonb WHERE state_id=$1",
		s.StateID, v, key)

	return err
}

func (s *state) deleteState(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM state WHERE state_id=$1", s.StateID)

	return err
}

func (s *state) deleteStateJSON(db *sql.DB, value string) error {
	_, err := db.Exec("UPDATE state SET data = data - $2 WHERE state_id=$1",
		s.StateID, value)

	return err
}
