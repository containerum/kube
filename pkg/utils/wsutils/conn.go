package wsutils

import (
	"encoding/json"

	"github.com/containerum/cherry"
	"github.com/gorilla/websocket"
)

func CloseWithCherry(conn *websocket.Conn, errToSend *cherry.Err) error {
	defer conn.Close()
	msg, _ := json.Marshal(errToSend)
	return conn.WriteMessage(websocket.CloseMessage, msg)
}
