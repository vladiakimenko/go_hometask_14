package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"urlshortener/internal/shortener"
)

// validators
func validateHeaders(t *testing.T, result *http.Response) {
	ctHeader := result.Header.Get("Content-Type")
	if ctHeader != "application/json" {
		t.Errorf("Некорректный 'Content-Type' хедер: %q", ctHeader)
	}
}

func validateStatus(t *testing.T, result *http.Response, wanted int) {
	status := result.StatusCode
	if status != wanted {
		t.Errorf("Ожидался статус %d, получено %d", wanted, status)
	}
}

func validateJsonResponse[T any](t *testing.T, result *http.Response) {
	bodyBytes, err := io.ReadAll(result.Body)
	if err != nil {
		t.Errorf("Не удалось прочитать тело ответа: %v", err)
		return
	}
	var rawMap map[string]json.RawMessage
	if err := json.Unmarshal(bodyBytes, &rawMap); err != nil {
		t.Errorf("Не удалось распарсить тело ответа как JSON: %v", err)
		return
	}

	typ := reflect.TypeFor[T]()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.Tag.Get("json") == "-" {
			continue
		}
		jsonKey := strings.Split(field.Tag.Get("json"), ",")[0]
		if jsonKey == "" {
			jsonKey = field.Name
		}
		if _, exists := rawMap[jsonKey]; !exists {
			t.Errorf("Поле %q отсутсвует в теле ответа", jsonKey)
		}
	}
}

func TestShorten(t *testing.T) {
	endpoint := "/shorten"
	tests := []struct {
		name          string
		requestURL    string
		requestMethod string
		requestBody   []byte
		wantStatus    int
	}{
		{"Валидный метод, тело и url", endpoint, http.MethodPost, []byte(`{"url":"http://example.com"}`), http.StatusCreated},
		// метод
		{"Неверный метод GET", endpoint, http.MethodGet, []byte(`{"url":"http://example.com"}`), http.StatusMethodNotAllowed},
		{"Неверный метод PUT", endpoint, http.MethodPut, []byte(`{"url":"http://example.com"}`), http.StatusMethodNotAllowed},
		{"Неверный метод PATCH", endpoint, http.MethodPatch, []byte(`{"url":"http://example.com"}`), http.StatusMethodNotAllowed},
		{"Неверный метод DELETE", endpoint, http.MethodDelete, []byte(`{"url":"http://example.com"}`), http.StatusMethodNotAllowed},
		{"Неверный метод HEAD", endpoint, http.MethodHead, []byte(`{"url":"http://example.com"}`), http.StatusMethodNotAllowed},
		{"Неверный метод OPTIONS", endpoint, http.MethodOptions, []byte(`{"url":"http://example.com"}`), http.StatusMethodNotAllowed},
		{"Неверный метод CONNECT", endpoint, http.MethodConnect, []byte(`{"url":"http://example.com"}`), http.StatusMethodNotAllowed},
		{"Неверный метод TRACE", endpoint, http.MethodTrace, []byte(`{"url":"http://example.com"}`), http.StatusMethodNotAllowed},
		// тело
		{"Тело не json", endpoint, http.MethodPost, []byte("plaintext"), http.StatusBadRequest},
		{"Тело пустое", endpoint, http.MethodPost, []byte{}, http.StatusBadRequest},
		{"Нет ключа 'url'", endpoint, http.MethodPost, []byte(`{}`), http.StatusBadRequest},
		// значение url
		{"Значение ключа 'url' не строка", endpoint, http.MethodPost, []byte(`{"url": null}`), http.StatusBadRequest},
		{"Значение ключа 'url' невалидный url", endpoint, http.MethodPost, []byte(`{"url": "not-a-url"}`), http.StatusBadRequest},
		{"Значение ключа 'url' пустая строка", endpoint, http.MethodPost, []byte(`{"url": ""}`), http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				us := shortener.NewURLShortener()
				router := NewRouter(us)
				request := httptest.NewRequest(tt.requestMethod, tt.requestURL, bytes.NewReader(tt.requestBody))
				recorder := httptest.NewRecorder()

				router.ServeHTTP(recorder, request)
				result := recorder.Result()
				defer result.Body.Close()
				success := result.StatusCode < http.StatusBadRequest

				validateHeaders(t, result)

				validateStatus(t, result, tt.wantStatus)

				if success {
					validateJsonResponse[ShortenResponse](t, result)
				} else {
					validateJsonResponse[ErrorResponse](t, result)
				}
			},
		)
	}
}

func TestRedirect(t *testing.T) {
	endpoint := "/"
	underlyingURL := "http://example.com"
	tests := []struct {
		name          string
		urlFactory    func(string) string
		requestMethod string
		wantStatus    int
	}{
		{"Верный метод, существующий ID", func(id string) string { return endpoint + id }, http.MethodGet, http.StatusFound},
		// метод
		{"Неверный метод POST", func(id string) string { return endpoint + id }, http.MethodPost, http.StatusMethodNotAllowed},
		{"Неверный метод PUT", func(id string) string { return endpoint + id }, http.MethodPut, http.StatusMethodNotAllowed},
		{"Неверный метод PATCH", func(id string) string { return endpoint + id }, http.MethodPut, http.StatusMethodNotAllowed},
		{"Неверный метод DELETE", func(id string) string { return endpoint + id }, http.MethodPut, http.StatusMethodNotAllowed},
		{"Неверный метод HEAD", func(id string) string { return endpoint + id }, http.MethodHead, http.StatusMethodNotAllowed},
		{"Неверный метод OPTIONS", func(id string) string { return endpoint + id }, http.MethodOptions, http.StatusMethodNotAllowed},
		{"Неверный метод CONNECT", func(id string) string { return endpoint + id }, http.MethodConnect, http.StatusMethodNotAllowed},
		{"Неверный метод TRACE", func(id string) string { return endpoint + id }, http.MethodTrace, http.StatusMethodNotAllowed},
		// ID
		{"Несуществующий ID", func(_ string) string { return endpoint + shortener.GenerateUrlId() }, http.MethodGet, http.StatusNotFound},
		{"ID с измененным регистром", func(id string) string { return endpoint + strings.ToUpper(id) }, http.MethodGet, http.StatusNotFound},
		{"Лишние фрагменты пути", func(id string) string { return endpoint + shortener.GenerateUrlId() + "/test" }, http.MethodGet, http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				us := shortener.NewURLShortener()
				existingID, err := us.Shorten(underlyingURL)
				if err != nil {
					t.Fatalf("setup failed: %v", err)
				}
				router := NewRouter(us)
				request := httptest.NewRequest(tt.requestMethod, tt.urlFactory(existingID), nil)
				recorder := httptest.NewRecorder()

				router.ServeHTTP(recorder, request)
				result := recorder.Result()
				defer result.Body.Close()
				success := result.StatusCode == http.StatusBadRequest

				validateHeaders(t, result)

				validateStatus(t, result, tt.wantStatus)

				if success {
					location := result.Header.Get("Location")
					if location != underlyingURL {
						t.Errorf("Неверный редирект-url в ответе: %q, должен быть: %q", location, underlyingURL)
					}
				}
			},
		)
	}
}
