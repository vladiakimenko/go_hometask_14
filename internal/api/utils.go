package api

import (
	"encoding/json"
	"log"
	"net/http"

	"urlshortener/internal/shortener"
)

func WriteError(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

func WriteJson(w http.ResponseWriter, status int, body any) {
	serialized, err := json.Marshal(body)
	if err != nil {
		log.Printf("Failed to serialize %v as json: %v", body, err)
		WriteError(w, http.StatusInternalServerError, "Response serialization error")
		return
	}
	w.WriteHeader(status)
	if _, err := w.Write(serialized); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}

func Chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

func NewRouter(us *shortener.URLShortener) http.Handler {
	mux := http.NewServeMux()

	mux.Handle(
		"/shorten",
		Chain(
			http.HandlerFunc(ShortenHandler),
			JSONContentTypeMiddleware,
			AllowedMethodsMiddleware(http.MethodPost),
			ParsedBodyMiddleware,
			ShortenerMiddleware(us),
		),
	)

	mux.Handle(
		"/",
		Chain(
			http.HandlerFunc(RedirectHandler),
			JSONContentTypeMiddleware,
			AllowedMethodsMiddleware(http.MethodGet),
			PathParamsMiddleware("/:shortID"),
			ShortenerMiddleware(us),
		),
	)

	return mux
}
