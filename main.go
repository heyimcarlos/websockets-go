package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"golang.org/x/net/websocket"
)

type Server struct {
	// we have a string instead of bool as the value, and have that string be the remote address of the client
	conns map[*websocket.Conn]bool
}

func NewServer() *Server {
	return &Server{
		conns: make(map[*websocket.Conn]bool),
	}
}

func (s *Server) readLoop(ws *websocket.Conn) {
	buf := make([]byte, 1024)
	for {
		// read the websocket msg into the first param, in this case the 1024 length buffer we created above.
		x, err := ws.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("print err: ", err)
			continue
		}
		// slices from 0 -> x
		msg := buf[:x]
		// fmt.Println(string(msg))
		// send a reply back
		// ws.Write([]byte("Hey there, welcome to the chat!"))

		s.broadcast(msg)
	}
}

func (s *Server) wsOrderbookHandler(ws *websocket.Conn) {
	fmt.Println("new incoming connection from client to order-book: ", ws.RemoteAddr())
	ticker := time.NewTicker(5 * time.Second)
	// defer is like a cleanup function
	defer ticker.Stop()

	// done := make(chan bool): In Go, we communicate between goroutines using channels. This line creates a new channel that can send and receive boolean values. It's used here to signal the goroutine to stop.
	done := make(chan bool)

	// when we use the keyword "go" a new goroutine start
	// a goroutine is a lightweight thread managed by the Go runtime, this is how go handles concurrent operations
	// go func is creating and starting an anonymous function as a goroutine
	go func() {
		for {
			// wait for the a channel to be ready, then run its logic, if multiple are ready, select one at random.
			select {
			case <-done:
				return
			case t := <-ticker.C:
				fmt.Println("Tick at: ", t)
				ws.Write([]byte("Order ready"))
			}
		}
	}()

    // stop ticker after 60 seconds
    time.Sleep(60 * time.Second)
    done <- true

}

func (s *Server) wsHandler(ws *websocket.Conn) {
	fmt.Println("new incoming connection from client: ", ws.RemoteAddr())

	// maps in golang are not concurrent, ideally we should use a mutex.
	s.conns[ws] = true
	fmt.Println("open conns: ", s.conns)

	s.readLoop(ws)
}

func (s *Server) broadcast(b []byte) {
	// iterate through all connections
	for ws := range s.conns {
		// this function is calling itself and received each websocket connection individually
		go func(ws *websocket.Conn) {
			if _, err := ws.Write(b); err != nil {
				fmt.Println("write error: ", err)
			}
			// pass the current conn into the go func
		}(ws)
	}
}

func main() {
	server := NewServer()
	fmt.Println(server)
	http.Handle("/ws", websocket.Handler(server.wsHandler))
	http.Handle("/ws/subscribe/order-book", websocket.Handler(server.wsOrderbookHandler))
	// http.Handle("/broadcast", websocket.Handler(server.broadcast))
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		log.Fatalln(err)
	}
}
