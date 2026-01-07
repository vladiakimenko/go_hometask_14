package shortener

import (
	"crypto/rand"
	"fmt"
	"log"
	"regexp"
	"sync"
)

// gloabals
const shortIDLength int = 8

var urlRE = regexp.MustCompile(`^https?://[a-zA-Z0-9.-]+(:[0-9]+)?(/[\w./?%&=-]*)?$`)

// core
type URLShortener struct {
	urls map[string]string
	mu   sync.RWMutex
}

func NewURLShortener() *URLShortener {
	return &URLShortener{
		urls: make(map[string]string),
	}
}

func (us *URLShortener) Shorten(originalURL string) (string, error) {
	if !isValidURL(originalURL) {
		return "", fmt.Errorf("Url validation failed")
	}

	var shortID string
	us.mu.Lock()
	for {
		shortID = GenerateUrlId()
		_, exists := us.urls[shortID]
		if exists {
			log.Println("short_url collision, consider rising 'shortIDLength'")
			continue
		}
		break
	}

	us.urls[shortID] = originalURL
	us.mu.Unlock()

	return shortID, nil
}

func (us *URLShortener) GetOriginal(shortID string) (string, error) {
	us.mu.RLock()
	originalURL, exists := us.urls[shortID]
	us.mu.RUnlock()
	if !exists {
		return "", fmt.Errorf("Record with shortID '%s' does not exist", shortID)
	}
	return originalURL, nil
}

// helpers
func GenerateUrlId() string {
	b := make([]byte, (shortIDLength+1)/2)
	rand.Read(b)
	var result = fmt.Sprintf("%x", b)[:shortIDLength]
	return result
}

func isValidURL(str string) bool {
	return urlRE.MatchString(str)
}
