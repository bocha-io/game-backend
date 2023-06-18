package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bocha-io/game-backend/x/cors"
	"github.com/bocha-io/garnet/x/logger"
	"github.com/gorilla/mux"
)

func PingRoute(router *mux.Router) {
	router.HandleFunc("/ping", RegisterPing).Methods("GET", "POST", "OPTIONS")
}

func SendInternalErrorResponse(msg string, w *http.ResponseWriter) {
	(*w).WriteHeader(http.StatusInternalServerError)
	fmt.Fprint(*w, msg)
}

func SendBadRequestResponse(msg string, w *http.ResponseWriter) {
	(*w).WriteHeader(http.StatusBadRequest)
	fmt.Fprint(*w, msg)
}

func SendJSONResponse(message interface{}, w *http.ResponseWriter) error {
	v, err := json.Marshal(message)
	if err != nil {
		SendInternalErrorResponse("invalid encoding for response", w)
		return err
	}
	(*w).Header().Set("Content-Type", "application/json")
	(*w).WriteHeader(http.StatusOK)
	_, err = (*w).Write(v)
	return err
}

func RegisterPing(response http.ResponseWriter, request *http.Request) {
	if cors.SetHandlerCorsForOptions(request, &response) {
		return
	}
	response.WriteHeader(http.StatusOK)
	_, err := response.Write([]byte("pong"))
	if err != nil {
		logger.LogDebug(fmt.Sprintf("[api] error sending pong: %s", err.Error()))
	}
}
