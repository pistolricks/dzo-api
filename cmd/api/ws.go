package main

import (
	"github.com/gobwas/ws"
	"github.com/mailru/easygo/netpoll"
	gopool "github.com/pistolricks/go-api-template/internal/pool"
	iws "github.com/pistolricks/go-api-template/internal/ws"
	"log"
	"net"
	"net/http"
	"time"
)

func (app *application) websockets() error {

	if x := app.config.ws.debug; x != "" {
		app.logger.Info("starting pprof server on %s", x, "pprof server error: %v", http.ListenAndServe(x, nil))
	}

	// Initialize netpoll instance. We will use it to be noticed about incoming
	// events from listener of user connections.
	poller, err := netpoll.New(nil)
	if err != nil {
		return err
	}

	var (
		// Make pool of X size, Y sized work queue and one pre-spawned
		// goroutine.
		pool    = gopool.NewPool(app.config.ws.workers, app.config.ws.queue, 1)
		message = iws.NewMessage(pool)
		exit    = make(chan struct{})
	)
	// handle is a new incoming messageion handler.
	// It upgrades TCP connection to WebSocket, registers netpoll listener on
	// it and stores it as a chat user in Chat instance.
	//
	// We will call it below within accept() loop.
	handle := func(conn net.Conn) {
		// NOTE: we wrap conn here to show that ws could work with any kind of
		// io.ReadWriter.
		safeConn := deadliner{conn, app.config.ws.ioTimeout}

		// Zero-copy upgrade to WebSocket connection.
		hs, err := ws.Upgrade(safeConn)
		if err != nil {

			app.logger.Info("%s: upgrade error: %v", nameConn(conn), err)

			err := conn.Close()
			if err != nil {
				return
			}
			return
		}
		app.logger.Info("%s: established websocket connection: %+v", nameConn(conn), hs)

		// Register incoming user in chat.
		agent := message.Register(safeConn)

		// Create netpoll event descriptor for conn.
		// We want to handle only read events of it.
		desc := netpoll.Must(netpoll.HandleRead(conn))

		// Subscribe to events about conn.
		err = poller.Start(desc, func(ev netpoll.Event) {
			if err != nil {
				return
			}
			if ev&(netpoll.EventReadHup|netpoll.EventHup) != 0 {
				// When ReadHup or Hup received, this mean that client has
				// closed at least write end of the connection or connections
				// itself. So we want to stop receive events about such conn
				// and remove it from the chat registry.
				err := poller.Stop(desc)
				if err != nil {
					return
				}
				message.Remove(agent)
				return
			}
			// Here we can read some new message from connection.
			// We can not read it right here in callback, because then we will
			// block the poller's inner loop.
			// We do not want to spawn a new goroutine to read single message.
			// But we want to reuse previously spawned goroutine.
			pool.Schedule(func() {
				if err := agent.Receive(); err != nil {
					// When receive failed, we can only disconnect broken
					// connection and stop to receive events about it.
					err := poller.Stop(desc)
					if err != nil {
						return
					}
					message.Remove(agent)
				}
			})
		})
		if err != nil {
			return
		}
	}

	// Create incoming connections listener.
	ln, err := net.Listen("tcp", app.config.ws.addr)
	if err != nil {
		return err
	}
	app.logger.Info("websocket is listening on %s", ln.Addr().String())

	// Create netpoll descriptor for the listener.
	// We use OneShot here to manually resume events stream when we want to.
	acceptDesc := netpoll.Must(netpoll.HandleListener(
		ln, netpoll.EventRead|netpoll.EventOneShot,
	))

	// accept is a channel to signal about next incoming connection Accept()
	// results.
	accept := make(chan error, 1)

	// Subscribe to events about listener.
	err = poller.Start(acceptDesc, func(e netpoll.Event) {
		// We do not want to accept incoming connection when goroutine pool is
		// busy. So if there are no free goroutines during 1ms we want to
		// cooldown the server and do not receive connection for some short
		// time.
		err := pool.ScheduleTimeout(time.Millisecond, func() {
			conn, err := ln.Accept()
			if err != nil {
				accept <- err
				return
			}

			accept <- nil
			handle(conn)
		})
		if err == nil {
			err = <-accept
		}
		if err != nil {
			if err != gopool.ErrScheduleTimeout {
				goto cooldown
			}
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				goto cooldown
			}

			log.Fatalf("accept error: %v", err)

		cooldown:
			delay := 5 * time.Millisecond
			log.Printf("accept error: %v; retrying in %s", err, delay)
			time.Sleep(delay)
		}

		err = poller.Resume(acceptDesc)
		if err != nil {
			return
		}
	})
	if err != nil {
		return err
	}

	<-exit
	return nil
}

func nameConn(conn net.Conn) string {
	return conn.LocalAddr().String() + " > " + conn.RemoteAddr().String()
}

// deadliner is a wrapper around net.Conn that sets read/write deadlines before
// every Read() or Write() call.
type deadliner struct {
	net.Conn
	t time.Duration
}

func (d deadliner) Write(p []byte) (int, error) {
	if err := d.Conn.SetWriteDeadline(time.Now().Add(d.t)); err != nil {
		return 0, err
	}
	return d.Conn.Write(p)
}

func (d deadliner) Read(p []byte) (int, error) {
	if err := d.Conn.SetReadDeadline(time.Now().Add(d.t)); err != nil {
		return 0, err
	}
	return d.Conn.Read(p)
}
