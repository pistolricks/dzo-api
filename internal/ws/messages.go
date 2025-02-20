package ws

import (
	"bytes"
	"encoding/json"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/pistolricks/go-api-template/internal/pool"
	"math/rand"
	"net"
	"sort"
	"strconv"
	"sync"
	"time"
)

type Message struct {
	mu    sync.RWMutex
	seq   int64
	agent []*Agent
	ns    map[string]*Agent

	pool *gopool.Pool
	out  chan []byte
}

func NewMessage(pool *gopool.Pool) *Message {
	message := &Message{
		pool: pool,
		ns:   make(map[string]*Agent),
		out:  make(chan []byte, 1),
	}

	go message.writer()

	return message
}

// Register registers new connection as a Agent.
func (m *Message) Register(conn net.Conn) *Agent {

	agent := &Agent{
		message: m,
		conn:    conn,
	}

	m.mu.Lock()
	{
		agent.id = m.seq
		agent.name = m.randName()

		m.agent = append(m.agent, agent)
		m.ns[agent.name] = agent

		m.seq++
	}
	m.mu.Unlock()

	err := agent.writeNotice("hello", Object{
		"name": agent.name,
	})
	if err != nil {
		return nil
	}
	err = m.Broadcast("greet", Object{
		"name": agent.name,
		"time": timestamp(),
	})
	if err != nil {
		return nil
	}

	return agent
}

// Remove removes agent from message.
func (m *Message) Remove(agent *Agent) {
	m.mu.Lock()
	removed := m.remove(agent)
	m.mu.Unlock()

	if !removed {
		return
	}

	err := m.Broadcast("goodbye", Object{
		"name": agent.name,
		"time": timestamp(),
	})
	if err != nil {
		return
	}
}

// Rename renames agent.
func (m *Message) Rename(agent *Agent, name string) (prev string, ok bool) {
	m.mu.Lock()
	{
		if _, has := m.ns[name]; !has {
			ok = true
			prev, agent.name = agent.name, name
			delete(m.ns, prev)
			m.ns[name] = agent
		}
	}
	m.mu.Unlock()

	return prev, ok
}

// Broadcast sends message to all alive agents.
func (m *Message) Broadcast(method string, params Object) error {
	var buf bytes.Buffer

	w := wsutil.NewWriter(&buf, ws.StateServerSide, ws.OpText)
	encoder := json.NewEncoder(w)

	r := Request{Method: method, Params: params}
	if err := encoder.Encode(r); err != nil {
		return err
	}
	if err := w.Flush(); err != nil {
		return err
	}

	m.out <- buf.Bytes()

	return nil
}

// writer writes broadcast messages from message.out channel.
func (m *Message) writer() {
	for bts := range m.out {
		m.mu.RLock()
		us := m.agent
		m.mu.RUnlock()

		for _, u := range us {
			u := u // For closure.
			m.pool.Schedule(func() {
				err := u.writeRaw(bts)
				if err != nil {
					return
				}
			})
		}
	}
}

// mutex must be held.
func (m *Message) remove(agent *Agent) bool {
	if _, has := m.ns[agent.name]; !has {
		return false
	}

	delete(m.ns, agent.name)

	i := sort.Search(len(m.agent), func(i int) bool {
		return m.agent[i].id >= agent.id
	})
	if i >= len(m.agent) {
		panic("message: inconsistent state")
	}

	without := make([]*Agent, len(m.agent)-1)
	copy(without[:i], m.agent[:i])
	copy(without[i:], m.agent[i+1:])
	m.agent = without

	return true
}

func (m *Message) randName() string {
	var suffix string
	for {
		name := animals[rand.Intn(len(animals))] + suffix
		if _, has := m.ns[name]; !has {
			return name
		}
		suffix += strconv.Itoa(rand.Intn(10))
	}

}

func timestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
