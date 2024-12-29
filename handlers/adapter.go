package handlers

import (
	"log/slog"
	"net/http"
)

type MatchError func(err error) bool

type HandleError func(err error, w http.ResponseWriter, r *http.Request)

type Handle func(w http.ResponseWriter, r *http.Request) error

type HandlerAdapter struct {
	pathes        []string
	handlers      []Handle
	errorHandlers []errorCaseHandler
}

func NewHandlerAdapter(pathes ...string) *HandlerAdapter {
	return &HandlerAdapter{
		pathes:        pathes,
		errorHandlers: []errorCaseHandler{},
		handlers:      []Handle{},
	}

}

func (h HandlerAdapter) ServeMux(mux *http.ServeMux) {
	for _, path := range h.pathes {
		mux.Handle(path, h)
	}
}

func (h *HandlerAdapter) AddHandler(handle Handle) {
	h.handlers = append(h.handlers, handle)
}

func (h *HandlerAdapter) AddErrorHandler(match MatchError, handle HandleError) {
	handler := errorCaseHandler{match: match, handle: handle}
	h.errorHandlers = append(h.errorHandlers, handler)
}

func (h HandlerAdapter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, handle := range h.handlers {
		if err := handle(w, r); err != nil {
			if !h.serveError(err, w, r) {
				slog.Error("ServeHTTP Unhandled Error", "error", err, "url", r.URL)
				break
			}
		}
	}
}

func (h HandlerAdapter) serveError(err error, w http.ResponseWriter, r *http.Request) bool {
	for _, handleError := range h.errorHandlers {
		if handleError.match(err) {
			handleError.handle(err, w, r)
			return true
		}
	}
	return false
}

type errorCaseHandler struct {
	match  MatchError
	handle HandleError
}
