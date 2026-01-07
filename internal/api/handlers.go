package api

import (
	"fmt"
	"log"
	"net/http"

	"urlshortener/internal/shortener"
)

func ShortenHandler(w http.ResponseWriter, r *http.Request) {
	body, ok := r.Context().Value(parsedBodyKey).(map[string]any)
	if !ok {
		log.Println("Request body was never parsed")
		WriteError(w, http.StatusInternalServerError, "Misconfigured handler")
		return
	}
	url, exists := body["url"]
	if !exists {
		WriteError(w, http.StatusBadRequest, "Missing key 'url' in request")
		return
	}
	urlStr, ok := url.(string)
	if !ok {
		WriteError(w, http.StatusBadRequest, "'url' must be a string")
		return
	}

	shortener, ok := r.Context().Value(shortenerKey).(*shortener.URLShortener)
	if !ok {
		log.Println("Shortener global instance was never attached to the context")
		WriteError(w, http.StatusInternalServerError, "Misconfigured handler")
		return
	}
	urlID, err := shortener.Shorten(urlStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, fmt.Sprintf("Shortening failed: %v", err))
		return
	}

	WriteJson(w, http.StatusCreated, ShortenResponse{Url: urlStr, ShortUrl: urlID})
}

func RedirectHandler(w http.ResponseWriter, r *http.Request) {
	pathParams, ok := r.Context().Value(pathParamsKey).(map[string]string)
	shortID, exists := pathParams["shortID"]
	if !ok || !exists {
		log.Printf("Path params parsing misconfigured for %s %s", r.Method, r.URL.Path)
		WriteError(w, http.StatusInternalServerError, "Misconfigured handler")
		return
	}

	shortener, ok := r.Context().Value(shortenerKey).(*shortener.URLShortener)
	if !ok {
		log.Println("Shortener global instance was never attached to the context")
		WriteError(w, http.StatusInternalServerError, "Misconfigured handler")
		return
	}
	originalUrl, err := shortener.GetOriginal(shortID)
	if err != nil {
		WriteError(w, http.StatusNotFound, "Not found")
		return
	}

	http.Redirect(w, r, originalUrl, http.StatusFound)
}
