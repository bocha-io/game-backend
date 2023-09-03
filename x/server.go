package backend

import (
	"fmt"
	"net/http"
	"time"

	"github.com/bocha-io/game-backend/x/api"
	"github.com/bocha-io/game-backend/x/cors"
	"github.com/bocha-io/game-backend/x/messages"
	"github.com/bocha-io/garnet/x/indexer/data"
	"github.com/bocha-io/logger"
	"github.com/gorilla/mux"
)

func NewGorillaServer(
	port int,
	mudDatabase *data.Database,
	handleMessage func(g *messages.Server, ws *messages.WebSocketContainer, m messages.BasicMessage, p []byte) error,
	handleDisconnect func(ws *messages.WebSocketContainer),
) (*messages.Server, *http.Server) {
	logger.LogInfo(fmt.Sprintf("[backend] starting server at port: %d\n", port))
	router := mux.NewRouter()

	g := messages.NewServer(mudDatabase, handleMessage, handleDisconnect)
	router.HandleFunc("/ws", g.WebSocketConnectionHandler).Methods("GET", "OPTIONS")
	api.PingRoute(router)

	cors.ServerEnableCORS(router)

	server := &http.Server{
		Addr:              fmt.Sprint(":", port),
		Handler:           router,
		ReadHeaderTimeout: 3 * time.Second,
	}
	return g, server
}

func StartGorillaServer(
	port int,
	mudDatabase *data.Database,
	handleMessage func(g *messages.Server, ws *messages.WebSocketContainer, m messages.BasicMessage, p []byte) error,
	handleDisconnect func(ws *messages.WebSocketContainer),
) error {
	_, server := NewGorillaServer(port, mudDatabase, handleMessage, handleDisconnect)
	return server.ListenAndServe()
}
