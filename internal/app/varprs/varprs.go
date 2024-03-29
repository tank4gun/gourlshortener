package varprs

import (
	"encoding/json"
	"flag"
	"os"
)

// FileStoragePath - path to the file storage
var FileStoragePath string

// BaseURL - base URL for shorten URLs, i.e. http://localhost:8080
var BaseURL string

// ServerAddress - address for running URLShortener app
var ServerAddress string

// GRPCServerAddress - address for running URLShortener app in GRPC mode
var GRPCServerAddress string

// DatabaseDSN - database connection address
var DatabaseDSN string

// UseHTTPS - flag for HTTPS enabling
var UseHTTPS bool

// ConfigPath - path to config with environment variables
var ConfigPath string

// TrustedSubnet - subnet mask
var TrustedSubnet string

// ConfigStruct - struct to parse config file
type ConfigStruct struct {
	ServerAddress     string `json:"server_address"`      // ServerAddress - server address for urlshortener app
	GRPCServerAddress string `json:"grpc_server_address"` // GRPCServerAddress - server address for urlshortener app in GRPC mode
	BaseURL           string `json:"base_url"`            // BaseURL - base route for URLs
	FileStoragePath   string `json:"file_storage_path"`   // FileStoragePath - path to file with data in case no db storage allowed
	DatabaseDSN       string `json:"database_dsn"`        // DatabaseDSN - connection string to database
	EnableHTTPS       bool   `json:"enable_https"`        // EnableHTTPS - flag in order to enable https
	TrustedSubnet     string `json:"trusted_subnet"`      // TrustedSubnet - flag for trusted subnet for handle GET /api/internal/stats
}

// ParseConfigFile - function got parsing conflict file
func ParseConfigFile() ConfigStruct {
	configPath := os.Getenv("CONFIG")
	if configPath != "" {
		ConfigPath = configPath
	}
	if ConfigPath == "" {
		return ConfigStruct{}
	}
	bytes, err := os.ReadFile(ConfigPath)
	if err != nil {
		return ConfigStruct{}
	}
	parsedConfig := ParseConfigFile
	err = json.Unmarshal(bytes, &parsedConfig)
	if err != nil {
		return ConfigStruct{}
	}
	return parsedConfig()
}

// Init - method for parsing environment variables and variables from configs
func Init() {
	flag.StringVar(&ServerAddress, "a", "", "Server address")
	flag.StringVar(&GRPCServerAddress, "gprc_addr", "", "GRPC server address")
	flag.StringVar(&BaseURL, "b", "", "Base URL for shorten URLs")
	flag.StringVar(&FileStoragePath, "f", "", "File path for storage")
	flag.StringVar(&DatabaseDSN, "d", "", "Database connection address")
	flag.BoolVar(&UseHTTPS, "s", false, "Use HTTPS for server")
	flag.StringVar(&ConfigPath, "config", "", "Config file path")
	flag.StringVar(&ConfigPath, "c", "", "Config file path")
	flag.StringVar(&TrustedSubnet, "t", "192.168.1.1/24", "Subnet mask")
	flag.Parse()

	config := ParseConfigFile()

	fileStoragePathEnv := os.Getenv("FILE_STORAGE_PATH")
	if fileStoragePathEnv != "" {
		FileStoragePath = fileStoragePathEnv
	}
	if FileStoragePath == "" {
		FileStoragePath = config.FileStoragePath
	}

	baseURLEnv := os.Getenv("BASE_URL")
	if baseURLEnv != "" {
		BaseURL = baseURLEnv
	} else {
		if BaseURL == "" {
			BaseURL = "http://localhost:8080"
		}
	}
	if BaseURL == "" {
		BaseURL = config.BaseURL
	}
	BaseURL += "/"

	serverAddrEnv := os.Getenv("SERVER_ADDRESS")
	if serverAddrEnv != "" {
		ServerAddress = serverAddrEnv
	} else {
		if ServerAddress == "" {
			ServerAddress = config.ServerAddress
		}
		if ServerAddress == "" {
			ServerAddress = "localhost:8080"
		}
	}

	grpcServerAddrEnv := os.Getenv("GRPC_SERVER_ADDRESS")
	if grpcServerAddrEnv != "" {
		GRPCServerAddress = grpcServerAddrEnv
	} else {
		if GRPCServerAddress == "" {
			GRPCServerAddress = config.GRPCServerAddress
		}
		if GRPCServerAddress == "" {
			GRPCServerAddress = "localhost:8081"
		}
	}

	databaseDSNEnv := os.Getenv("DATABASE_DSN")
	if databaseDSNEnv != "" {
		DatabaseDSN = databaseDSNEnv
	}
	if DatabaseDSN == "" {
		DatabaseDSN = config.DatabaseDSN
	}
	useHTTPS := os.Getenv("ENABLE_HTTPS")
	if useHTTPS != "" {
		UseHTTPS = true
	}
	if !UseHTTPS {
		UseHTTPS = config.EnableHTTPS
	}
	trustedSubnet := os.Getenv("TRUSTED_SUBNET")
	if trustedSubnet != "" {
		TrustedSubnet = trustedSubnet
	}
	if TrustedSubnet == "" {
		TrustedSubnet = config.TrustedSubnet
	}
}
