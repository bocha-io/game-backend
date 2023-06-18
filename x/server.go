package backend

import (
	"fmt"
	"net/http"
	"time"

	"github.com/bocha-io/game-backend/x/api"
	"github.com/bocha-io/game-backend/x/cors"
	"github.com/bocha-io/game-backend/x/messages"
	"github.com/bocha-io/garnet/x/indexer/data"
	"github.com/bocha-io/garnet/x/logger"
	"github.com/gorilla/mux"
)

func StartGorillaServer(
	port int,
	mudDatabase *data.Database,
	HandleMessage func(g *messages.Connection, ws *messages.WebSocketContainer, m messages.BasicMessage, p []byte) error,
) error {
	logger.LogInfo(fmt.Sprintf("[backend] starting server at port: %d\n", port))
	router := mux.NewRouter()

	g := messages.NewConnection(mudDatabase)
	router.HandleFunc("/ws", g.WebSocketConnectionHandler).Methods("GET", "OPTIONS")
	api.PingRoute(router)

	cors.ServerEnableCORS(router)

	server := &http.Server{
		Addr:              fmt.Sprint(":", port),
		Handler:           router,
		ReadHeaderTimeout: 3 * time.Second,
	}
	return server.ListenAndServe()
}
