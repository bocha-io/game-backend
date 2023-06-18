package messages

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/bocha-io/game-backend/x/cors"
	"github.com/bocha-io/garnet/x/indexer/data"
	"github.com/bocha-io/garnet/x/logger"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type WebSocketContainer struct {
	Authenticated bool
	User          string
	WalletID      int
	WalletAddress string
	Conn          *websocket.Conn
}

type Connection struct {
	done              chan (struct{})
	WalletIndex       map[string]string
	WsSockets         map[string]*WebSocketContainer
	Database          *data.Database
	LastBroadcastTime time.Time
	HandleMessage     func(g *Connection, ws *WebSocketContainer, m BasicMessage, p []byte) error
}

// NOTE: your connect msg should set the web socket container values
// ws.User = connectMsg.User
// ws.Authenticated = true
// ws.WalletID = user.Index
// ws.WalletAddress = strings.ToLower(user.Address)

func NewConnection(
	database *data.Database,
	HandleMessage func(g *Connection, ws *WebSocketContainer, m BasicMessage, p []byte) error,

) Connection {
	return Connection{
		done:              make(chan struct{}),
		WalletIndex:       make(map[string]string),
		WsSockets:         make(map[string]*WebSocketContainer),
		Database:          database,
		LastBroadcastTime: time.Now(),
	}
}

func (g *Connection) WebSocketConnectionHandler(response http.ResponseWriter, request *http.Request) {
	if cors.SetHandlerCorsForOptions(request, &response) {
		return
	}

	// TODO: Filter prod page or localhost for development
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	ws, err := upgrader.Upgrade(response, request, nil)
	if err != nil {
		// Maybe log the error
		return
	}

	webSocket := WebSocketContainer{
		Authenticated: false,
		Conn:          ws,
	}

	g.WsHandler(&webSocket)
}

func WriteMessage(ws *websocket.Conn, msg *string) error {
	return ws.WriteMessage(websocket.TextMessage, []byte(*msg))
}

func RemoveConnection(ws *WebSocketContainer, g *Connection) {
	ws.Conn.Close()
	delete(g.WsSockets, ws.User)
}

func (g *Connection) WsHandler(ws *WebSocketContainer) {
	for {
		defer RemoveConnection(ws, g)
		// Read until error the client messages
		_, p, err := ws.Conn.ReadMessage()
		if err != nil {
			return
		}

		logger.LogDebug(fmt.Sprintf("[backend] incoming message: %s", string(p)))

		var m BasicMessage
		err = json.Unmarshal(p, &m)
		if err != nil {
			return
		}

		g.HandleMessage(g, ws, m, p)
	}
}
