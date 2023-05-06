package handlers

import (
	"github.com/tank4gun/gourlshortener/internal/app/storage"
	"github.com/tank4gun/gourlshortener/internal/app/types"
)

// ICommonServer interface is used as facade
type ICommonServer interface {
	CreateShortURL(storage storage.IRepository, URL string, userID uint, baseURL string) (shortURL string, errorMessage string, errorCode int)                                             // CreateShortURL - converts URL to shorten one and saves into storage
	GetURLByID(storage storage.IRepository, shortURL string, userID uint) (originalURL string, errorCode int)                                                                              // GetURLByID - returns full URL by its ID if it exists in storage
	CreateShortenURLBatch(storage storage.IRepository, batchRequest []storage.BatchURLRequest, baseURL string) (resultURLs []storage.BatchURLResponse, errorMessage string, errorCode int) // CreateShortenURLBatch - converts URL batch to shorten one and saves into storage
	GetAllURLs(storage storage.IRepository, userID uint, baseURL string) (responseList []storage.FullInfoURLResponse, errorCode int)                                                       // GetAllURLs - return all URLs for given User from storage
	DeleteURLs(deleteChannel chan types.RequestToDelete, URLsToDelete []string, userID uint)                                                                                               // DeleteURLs - removes all URLs for given User from storage
	Ping(storage storage.IRepository) error                                                                                                                                                // Ping - checks than connection to storage is alive
	GetStats(storage storage.IRepository) (stats storage.StatsResponse, errorCode int)                                                                                                     // GetStats - gets statistics, return all URLs and Users number from storage
}

// CommonServer - implementation for ICommonServer
type CommonServer struct{}

// CreateShortURL - converts URL to shorten one and saves into storage
func (server CommonServer) CreateShortURL(storage storage.IRepository, URL string, userID uint, baseURL string) (shortURL string, errorMessage string, errorCode int) {
	shortURL, errorMessage, errorCode = storage.CreateShortURLByURL(URL, userID)
	return baseURL + shortURL, errorMessage, errorCode
}

// GetURLByID - returns full URL by its ID if it exists in storage
func (server CommonServer) GetURLByID(storage storage.IRepository, shortURL string, userID uint) (originalURL string, errorCode int) {
	id := ConvertShortURLToID(shortURL)
	originalURL, errorCode = storage.GetValueByKeyAndUserID(id, userID)
	return originalURL, errorCode
}

// CreateShortenURLBatch - converts URL batch to shorten one and saves into storage
func (server CommonServer) CreateShortenURLBatch(storage storage.IRepository, batchRequest []storage.BatchURLRequest, userID uint, baseURL string) (resultURLs []storage.BatchURLResponse, errorMessage string, errorCode int) {
	resultURLs, errorMessage, errorCode = storage.CreateShortURLBatch(batchRequest, userID, baseURL)
	return resultURLs, errorMessage, errorCode
}

// GetAllURLs - return all URLs for given User from storage
func (server CommonServer) GetAllURLs(storage storage.IRepository, userID uint, baseURL string) (responseList []storage.FullInfoURLResponse, errorCode int) {
	responseList, errorCode = storage.GetAllURLsByUserID(userID, baseURL)
	return responseList, errorCode
}

// DeleteURLs - removes all URLs for given User from storage
func (server CommonServer) DeleteURLs(deleteChannel chan types.RequestToDelete, URLsToDelete []string, userID uint) {
	go func() {
		deleteChannel <- types.RequestToDelete{URLs: URLsToDelete, UserID: userID}
	}()
}

// Ping - checks than connection to storage is alive
func (server CommonServer) Ping(storage storage.IRepository) error {
	return storage.Ping()
}

// GetStats - gets statistics, return all URLs and Users number from storage
func (server CommonServer) GetStats(storage storage.IRepository) (stats storage.StatsResponse, errorCode int) {
	stats, errorCode = storage.GetStats()
	return stats, errorCode
}
