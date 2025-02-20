package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
)

func (app *application) proxy() {
	flag.Parse()
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("can not get os working directory: %v", err)
	}

	web := http.FileServer(http.Dir(wd + "/web"))

	http.Handle("/", web)
	http.Handle("/web/", http.StripPrefix("/web/", web))
	http.Handle("/ws", app.upstream("message", "tcp", app.config.proxy.messageAddr))

	log.Printf("proxy is listening on %q", app.config.proxy.addr)
	log.Fatal(http.ListenAndServe(app.config.proxy.addr, nil))
}

func (app *application) upstream(name, network, addr string) http.Handler {
	if conn, err := net.Dial(network, addr); err != nil {
		log.Printf("warning: test upstream %q error: %v", name, err)
	} else {
		log.Printf("upstream %q ok", name)
		err := conn.Close()
		if err != nil {
			return nil
		}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		peer, err := net.Dial(network, addr)
		if err != nil {
			m := fmt.Sprintf("dial upstream error: %v", err)
			app.baseErrorResponse(w, r, http.StatusBadGateway, m, err)
			return
		}
		if err := r.Write(peer); err != nil {
			m := fmt.Sprintf("write request to upstream error: %v", err)
			app.baseErrorResponse(w, r, http.StatusBadGateway, m, err)
			return
		}
		hj, ok := w.(http.Hijacker)
		if !ok {
			app.serverErrorResponse(w, r, err)
			return
		}
		conn, _, err := hj.Hijack()
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		log.Printf(
			"serving %s < %s <~> %s > %s",
			peer.RemoteAddr(), peer.LocalAddr(), conn.RemoteAddr(), conn.LocalAddr(),
		)

		go func() {
			defer func(peer net.Conn) {
				err := peer.Close()
				if err != nil {
					app.serverErrorResponse(w, r, err)
				}
			}(peer)
			defer func(conn net.Conn) {
				err := conn.Close()
				if err != nil {
					app.serverErrorResponse(w, r, err)
				}
			}(conn)
			_, err := io.Copy(peer, conn)
			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}
		}()
		go func() {
			defer func(peer net.Conn) {
				err := peer.Close()
				if err != nil {
					app.serverErrorResponse(w, r, err)
				}
			}(peer)
			defer func(conn net.Conn) {
				err := conn.Close()
				if err != nil {
					app.serverErrorResponse(w, r, err)
				}
			}(conn)
			_, err := io.Copy(conn, peer)
			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}
		}()
	})
}

func indexHandler(wd string) (http.Handler, error) {
	index, err := os.Open(wd + "/web/index.html")
	if err != nil {
		return nil, err
	}
	stat, err := index.Stat()
	if err != nil {
		return nil, err
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeContent(w, r, "", stat.ModTime(), index)
	}), nil
}
