package shortener

import "testing"

func TestURLShortener_Shorten(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"валидный HTTP URL", "http://example.com", false},
		{"валидный HTTPS URL", "https://google.com/search?q=test", false},
		{"невалидный URL", "not-a-url", true},
		{"пустая строка", "", true},
	}
	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				shortener := NewURLShortener()
				shortID, err := shortener.Shorten(tt.url)
				if (err != nil) != tt.wantErr {
					t.Errorf("ошибка = %v, ожидали ошибку = %v", err, tt.wantErr)
					return
				}
				if !tt.wantErr && len(shortID) != shortIDLength {
					t.Errorf("короткий ID слишком короткий: %s", shortID)
				}
			},
		)
	}
}

func TestURLShortener_GetOriginal(t *testing.T) {
	lookupID := GenerateUrlId()
	underlyingURL := "https://example.com"
	tests := []struct {
		name         string
		shortID      string
		existingUrls map[string]string
		wantErr      bool
	}{
		{"ShortID существует", lookupID, map[string]string{lookupID: underlyingURL}, false},
		{"ShortID отсутствует", lookupID, map[string]string{}, true},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				shortener := NewURLShortener()
				shortener.urls = tt.existingUrls
				originalURL, err := shortener.GetOriginal(tt.shortID)
				if (err != nil) != tt.wantErr {
					t.Errorf("ошибка = %v, ожидали ошибку = %v", err, tt.wantErr)
					return
				}
				if !tt.wantErr && originalURL != underlyingURL {
					t.Errorf("вернулся неверный URL: %s", originalURL)
				}
			},
		)
	}
}
