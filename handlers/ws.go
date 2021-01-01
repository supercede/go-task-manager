package handlers

import (
	"net/http"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

var wss wsocket

type wsocket struct {
	Conn    *websocket.Conn
	Message message
}

type message struct {
	Username string `json:"username"`
	Action   string `json:"action"`
	Message  string `json:"message"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// func reader(conn *websocket.Conn) {
// 	for {
// 		// read in a message
// 		var msg message
// 		err := conn.ReadJSON(&msg)
// 		if err != nil {
// 			log.Info("read msg error: %v", err)
// 			break
// 		}
// 		// print out that message for clarity
// 		log.Info(msg)

// 		if err := conn.WriteJSON(msg); err != nil {
// 			log.Println(err)
// 			return
// 		}

// 	}
// }

func reader(ws *wsocket) {
	for {
		// read in a message
		err := ws.Conn.ReadJSON(&ws.Message)
		if err != nil {
			log.Info("read msg error: %v", err)
			break
		}
		// print out that message for clarity
		log.Info(ws.Message)

		if err := ws.Conn.WriteJSON(ws.Message); err != nil {
			log.Println(err)
			return
		}

	}
}

// There's a websocket.Conn pointer you get when you upgrade a request to websocket... You can store that in a map or something

func (h *Handler) WSEndpoint(w http.ResponseWriter, r *http.Request) {
	// upgrade this connection to a WebSocket
	// connection
	// var wss wsocket
	// ws := wss.Conn
	var err error
	wss.Conn, err = upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}

	log.Println("Client Connected")
	err = wss.Conn.WriteMessage(1, []byte("Hi Client!"))
	if err != nil {
		log.Println(err)
	}
	// listen indefinitely for new messages coming
	// through on our WebSocket connection
	reader(&wss)
}
