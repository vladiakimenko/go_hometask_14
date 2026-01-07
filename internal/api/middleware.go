package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"urlshortener/internal/shortener"
)

// globals
type contextKey string

const parsedBodyKey contextKey = "parsedBody"
const pathParamsKey contextKey = "pathParams"
const shortenerKey contextKey = "shortener"

// middleware
func JSONContentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			next.ServeHTTP(w, r)
		},
	)
}

func ParsedBodyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.Body != nil {
				defer r.Body.Close()
				var bodyMap map[string]any
				if err := json.NewDecoder(r.Body).Decode(&bodyMap); err != nil {
					WriteError(w, http.StatusBadRequest, "Could not parse request body as json")
					return
				}
				ctx := context.WithValue(r.Context(), parsedBodyKey, bodyMap)
				r = r.WithContext(ctx)
			}
			next.ServeHTTP(w, r)
		},
	)
}

func AllowedMethodsMiddleware(method string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				if r.Method != method {
					log.Printf("Method %s not allowed", r.Method)
					WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
					return
				}
				next.ServeHTTP(w, r)
			},
		)
	}
}

func PathParamsMiddleware(pattern string) func(http.Handler) http.Handler {
	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
				if len(pathParts) != len(patternParts) {
					WriteError(w, http.StatusNotFound, "Not found")
					return
				}
				params := map[string]string{}
				for i, part := range patternParts {
					if strings.HasPrefix(part, ":") {
						params[part[1:]] = pathParts[i]
					} else if part != pathParts[i] {
						WriteError(w, http.StatusNotFound, "Not found")
						return
					}
				}
				ctx := context.WithValue(r.Context(), pathParamsKey, params)
				next.ServeHTTP(w, r.WithContext(ctx))
			},
		)
	}
}

func ShortenerMiddleware(us *shortener.URLShortener) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				ctx := context.WithValue(r.Context(), shortenerKey, us)
				r = r.WithContext(ctx)
				next.ServeHTTP(w, r)
			},
		)
	}
}
