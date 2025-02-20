package ws

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type Object map[string]interface{}

type Error struct {
	ID    int    `json:"id"`
	Error Object `json:"error"`
}

type Request struct {
	ID     int    `json:"id"`
	Method string `json:"method"`
	Params Object `json:"params"`
}

type Response struct {
	ID     int    `json:"id"`
	Result Object `json:"result"`
}

type Ws struct {
	Agents AgentModel
}

func NewWs(db *sql.DB) Ws {
	return Ws{
		Agents: AgentModel{DB: db},
	}
}
