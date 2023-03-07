package varprs

import (
	"flag"
	"os"
)

// FileStoragePath - path to the file storage
var FileStoragePath string

// BaseURL - base URL for shorten URLs, i.e. http://localhost:8080
var BaseURL string

// ServerAddress - address for running URLShortener app
var ServerAddress string

// DatabaseDSN - database connection address
var DatabaseDSN string

// Init - method for parsing environment variables and variables from configs
func Init() {
	flag.StringVar(&ServerAddress, "a", "localhost:8080", "Server address")
	flag.StringVar(&BaseURL, "b", "http://localhost:8080", "Base URL for shorten URLs")
	flag.StringVar(&FileStoragePath, "f", "storage.txt", "File path for storage")
	flag.StringVar(&DatabaseDSN, "d", "", "Database connection address")
	flag.Parse()

	fileStoragePathEnv := os.Getenv("FILE_STORAGE_PATH")
	if fileStoragePathEnv != "" {
		FileStoragePath = fileStoragePathEnv
	}

	baseURLEnv := os.Getenv("BASE_URL")
	if baseURLEnv != "" {
		BaseURL = baseURLEnv
	} else {
		if BaseURL == "" {
			BaseURL = "http://localhost:8080"
		}
	}
	BaseURL += "/"

	serverAddrEnv := os.Getenv("SERVER_ADDRESS")
	if serverAddrEnv != "" {
		ServerAddress = serverAddrEnv
	} else {
		if ServerAddress == "" {
			ServerAddress = "localhost:8080"
		}
	}

	databaseDSNEnv := os.Getenv("DATABASE_DSN")
	if databaseDSNEnv != "" {
		DatabaseDSN = databaseDSNEnv
	}
}
