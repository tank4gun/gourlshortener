package types

// userCtxName - type string
type userCtxName string

// UserIDCtxName - context key for userID
var UserIDCtxName = userCtxName("UserID")

// CookieKey - key for cookie generator
var CookieKey = []byte("URL-Shortener-Key")

// URLShortenderCookieName - cookie name
var URLShortenderCookieName = "URL-Shortener"

// RequestToDelete - message type for URL deletion
type RequestToDelete struct {
	URLs   []string // URLs - list with URLs to delete
	UserID uint     // UserID - user ID for URLs to delete
}

// URLBodyRequest is a base structure for request
type URLBodyRequest struct {
	// URL to shorten
	URL string `json:"url"`
}

// ShortenURLResponse response for shorten URL creation
type ShortenURLResponse struct {
	// URL result
	URL string `json:"result"`
}
