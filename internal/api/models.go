package api

type ShortenResponse struct {
	Url      string `json:"url"`
	ShortUrl string `json:"short_url"`
}

type ErrorResponse = struct {
	Error string `json:"error"`
}
