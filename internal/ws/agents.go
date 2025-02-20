package ws

import (
	"database/sql"
	"encoding/json"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	_ "github.com/pistolricks/go-api-template/internal/pool"
	"io"
	"sync"
)

type Agent struct {
	io      sync.Mutex
	conn    io.ReadWriteCloser
	id      int64
	name    string
	message *Message
}

type AgentModel struct {
	DB *sql.DB
}

func (a *Agent) Receive() error {
	req, err := a.readRequest()
	if err != nil {
		err := a.conn.Close()
		if err != nil {
			return err
		}
		return err
	}
	if req == nil {
		// Handled some control message.
		return nil
	}
	switch req.Method {
	case "rename":
		name, ok := req.Params["name"].(string)
		if !ok {
			return a.writeErrorTo(req, Object{
				"error": "bad params",
			})
		}
		prev, ok := a.message.Rename(a, name)
		if !ok {
			return a.writeErrorTo(req, Object{
				"error": "already exists",
			})
		}
		err := a.message.Broadcast("rename", Object{
			"prev": prev,
			"name": name,
			"time": timestamp(),
		})
		if err != nil {
			return err
		}
		return a.writeResultTo(req, nil)
	case "publish":
		req.Params["author"] = a.name
		req.Params["time"] = timestamp()
		err := a.message.Broadcast("publish", req.Params)
		if err != nil {
			return err
		}
	default:
		return a.writeErrorTo(req, Object{
			"error": "not implemented",
		})
	}
	return nil
}

// readRequests reads json-rpc request from connection.
// It takes io mutex.
func (a *Agent) readRequest() (*Request, error) {
	a.io.Lock()
	defer a.io.Unlock()

	h, r, err := wsutil.NextReader(a.conn, ws.StateServerSide)
	if err != nil {
		return nil, err
	}
	if h.OpCode.IsControl() {
		return nil, wsutil.ControlFrameHandler(a.conn, ws.StateServerSide)(h, r)
	}

	req := &Request{}
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(req); err != nil {
		return nil, err
	}

	return req, nil
}

func (a *Agent) writeErrorTo(req *Request, err Object) error {
	return a.write(Error{
		ID:    req.ID,
		Error: err,
	})
}

func (a *Agent) writeResultTo(req *Request, result Object) error {
	return a.write(Response{
		ID:     req.ID,
		Result: result,
	})
}

func (a *Agent) writeNotice(method string, params Object) error {
	return a.write(Request{
		Method: method,
		Params: params,
	})
}

func (a *Agent) write(x interface{}) error {
	w := wsutil.NewWriter(a.conn, ws.StateServerSide, ws.OpText)
	encoder := json.NewEncoder(w)

	a.io.Lock()
	defer a.io.Unlock()

	if err := encoder.Encode(x); err != nil {
		return err
	}

	return w.Flush()
}

func (a *Agent) writeRaw(p []byte) error {
	a.io.Lock()
	defer a.io.Unlock()

	_, err := a.conn.Write(p)

	return err
}
