// Created by Eunan Camilleri ( eunancamilleri@gmail.com )

package main

import (
	"fmt"
	"github.com/eunanibus/fyusion-app/internal/app/fyusion-app/routes"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net/http"
)

const port = 8888

func main() {
	router := mux.NewRouter()
	handler := routes.NewRequestHandler()
	router.HandleFunc("/", handler.MediaUpload).Queries("callback", "{callback}").Methods(http.MethodPut)
	router.HandleFunc("/callback", handler.Callback).Methods(http.MethodPost)
	log.Infof("Server started. Listening on port %d", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), router))
}
