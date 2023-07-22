package messages

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/bocha-io/game-backend/x/cors"
	"github.com/bocha-io/garnet/x/indexer/data"
	"github.com/bocha-io/logger"
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
	ConnMutex     *sync.Mutex
}

type Server struct {
	done              chan (struct{})
	WalletIndex       map[string]string
	WsSockets         map[string]*WebSocketContainer
	Database          *data.Database
	LastBroadcastTime time.Time
	HandleMessage     func(g *Server, ws *WebSocketContainer, m BasicMessage, p []byte) error
}

// NOTE: your connect msg should set the web socket container values
// ws.User = connectMsg.User
// ws.Authenticated = true
// ws.WalletID = user.Index
// ws.WalletAddress = strings.ToLower(user.Address)

func NewServer(
	database *data.Database,
	handleMessage func(g *Server, ws *WebSocketContainer, m BasicMessage, p []byte) error,
) *Server {
	return &Server{
		done:              make(chan struct{}),
		WalletIndex:       make(map[string]string),
		WsSockets:         make(map[string]*WebSocketContainer),
		Database:          database,
		LastBroadcastTime: time.Now(),
		HandleMessage:     handleMessage,
	}
}

func (g *Server) WebSocketConnectionHandler(response http.ResponseWriter, request *http.Request) {
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
		ConnMutex:     &sync.Mutex{},
	}

	g.WsHandler(&webSocket)
}

// Probably make this function part of the WebSocketContainer to ignore the first 2 params
func WriteMessage(ws *websocket.Conn, wsMutex *sync.Mutex, msg *string) error {
	wsMutex.Lock()
	defer wsMutex.Unlock()

	return ws.WriteMessage(websocket.TextMessage, []byte(*msg))
}

// Probably make this function part of the WebSocketContainer to ignore the first 2 params
func WriteJSON(ws *websocket.Conn, wsMutex *sync.Mutex, msg interface{}) error {
	wsMutex.Lock()
	defer wsMutex.Unlock()

	value, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return ws.WriteMessage(websocket.TextMessage, value)
}

func RemoveConnection(ws *WebSocketContainer, g *Server) {
	// TODO: this should inform that the connection is closed so we do not broadcast to this client
	ws.Conn.Close()
	delete(g.WsSockets, ws.User)
}

func (g *Server) WsHandler(ws *WebSocketContainer) {
	defer RemoveConnection(ws, g)
	for {
		// Read until error the client messages
		_, p, err := ws.Conn.ReadMessage()
		if err != nil {
			return
		}

		logger.LogDebug(fmt.Sprintf("[backend] incoming message: %s", string(p)))

		var m BasicMessage
		err = json.Unmarshal(p, &m)
		if err != nil {
			logger.LogInfo("[backend] closing connection because the msg is not a basic message")
			return
		}

		if err := g.HandleMessage(g, ws, m, p); err != nil {
			logger.LogInfo(fmt.Sprintf("[backend] closing connection error in the handle message function: %s", err.Error()))
			return
		}
	}
}
