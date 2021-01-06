package ws

import (
	"net/url"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

// Connect serves as a client to the websocket endpoint
func Connect() (*websocket.Conn, error) {
	// Initiate the websocket connection from the go code **as a client** to connect to the ws endpoint
	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/ws"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}

	return c, nil
}
