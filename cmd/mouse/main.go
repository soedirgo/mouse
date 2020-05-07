// Server-side part of the Go websocket sample.
//
// Eli Bendersky [http://eli.thegreenplace.net]
// Bobbie Soedirgo [https://soedirgo.dev]
// This code is in the public domain.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"golang.org/x/net/trace"
	"golang.org/x/net/websocket"
)

var (
	port    = flag.Int("port", 4050, "The server port")
	cursors sync.Map
)

// Event represents a pointer position.
//
// The fields of this struct must be exported so that the json module will be
// able to write into them. Therefore we need field tags to specify the names
// by which these fields go in the JSON representation of events.
type Event struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// handleWebsocketEchoMessage handles the message e arriving on connection ws
// from the client.
func handleWebsocketEchoMessage(ws *websocket.Conn, e Event) error {
	// Log the request with net.Trace
	tr := trace.New("websocket.Receive", "receive")
	defer tr.Finish()
	tr.LazyPrintf("Got event %v\n", e)

	// Echo the event back as JSON
	err := websocket.JSON.Send(ws, e)
	if err != nil {
		return fmt.Errorf("Can't send: %s", err.Error())
	}
	cursors.Store(ws, e)
	return nil
}

// websocketEchoConnection handles a single websocket echo connection - ws.
func websocketEchoConnection(ws *websocket.Conn) {
	log.Printf("Client connected from %s", ws.RemoteAddr())
	for {
		var event Event
		err := websocket.JSON.Receive(ws, &event)
		if err != nil {
			log.Printf("Receive failed: %s; closing connection...", err.Error())
			cursors.Delete(ws)
			if err = ws.Close(); err != nil {
				log.Println("Error closing connection:", err.Error())
			}
			break
		} else {
			if err := handleWebsocketEchoMessage(ws, event); err != nil {
				log.Println(err.Error())
				break
			}
		}
	}
}

// websocketTimeConnection handles a single websocket time connection - ws.
func websocketTimeConnection(ws *websocket.Conn) {
	for range time.Tick(10 * time.Millisecond) {
		// Once a second, send a message (as a string) with the current time.
		values := struct {
			Players int `json:"players"`
			Cursors []Event `json:"cursors"`
		}{}
		cursors.Range(func(k, v interface{}) bool {
			values.Players++
			values.Cursors = append(values.Cursors, v.(Event))
			return true
		})
		websocket.JSON.Send(ws, values)
	}
}

func main() {
	flag.Parse()
	// Set up websocket servers and static file server. In addition, we're using
	// net/trace for debugging - it will be available at /debug/requests.
	http.Handle("/wsecho", websocket.Handler(websocketEchoConnection))
	http.Handle("/wstime", websocket.Handler(websocketTimeConnection))
	http.Handle("/", http.FileServer(http.Dir("web/static")))

	log.Printf("Server listening on port %d", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}
