package routes

import (
	"github.com/eunanibus/fyusion-app/internal/app/fyusion-app/handlers/demo"
	"github.com/eunanibus/fyusion-app/internal/app/fyusion-app/handlers/media"
	"net/http"
)

type Handler interface {
	MediaUpload(res http.ResponseWriter, req *http.Request)
	Callback(res http.ResponseWriter, req *http.Request)
}

type RequestHandler struct{}

func NewRequestHandler() *RequestHandler {
	return &RequestHandler{}
}

func (dh *RequestHandler) MediaUpload(res http.ResponseWriter, req *http.Request) {
	media.HandleMediaUpload(res, req)
}

func (dh *RequestHandler) Callback(res http.ResponseWriter, req *http.Request) {
	demo.DemoCallbackHandler(res, req)
}
