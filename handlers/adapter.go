package handlers

import (
	"log/slog"
	"net/http"
)

// MatchError defines a function that determines if an error mathces
// specific criteria.
type MatchError func(err error) bool

// HandleError defines a function to process errors during request handling.
type HandleError func(err error, w http.ResponseWriter, r *http.Request)

// Handle defines a function for handling HTTP requests.
type Handle func(w http.ResponseWriter, r *http.Request) error

// HandlerAdapter routes and handles HTTP requests with error handling capabilities.
type HandlerAdapter struct {
	paths         []string
	handlers      []Handle
	errorHandlers []errorCaseHandler
}

// NewHandlerAdapter initializes a new HandlerAdapter with the given paths.
func NewHandlerAdapter(paths ...string) *HandlerAdapter {
	return &HandlerAdapter{
		paths:         paths,
		errorHandlers: []errorCaseHandler{},
		handlers:      []Handle{},
	}
}

func (h HandlerAdapter) Paths() []string {
	copied := make([]string, len(h.paths))
	copy(copied, h.paths)
	return copied
}

// ServeMux registers the HandlerAdapter's paths with the given ServeMux.
func (h HandlerAdapter) ServeMux(mux *http.ServeMux) {
	for _, path := range h.paths {
		mux.Handle(path, h)
	}
}

// AddHandler appends a new request handler to the HandlerAdapter.
func (h *HandlerAdapter) AddHandler(handle Handle) {
	h.handlers = append(h.handlers, handle)
}

// AddErrorHandler registers a new error handler with matching criteria.
func (h *HandlerAdapter) AddErrorHandler(match MatchError, handle HandleError) {
	handler := errorCaseHandler{match: match, handle: handle}
	h.errorHandlers = append(h.errorHandlers, handler)
}

// ServeHTTP processes incoming HTTP requests and handles errors as needed.
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
