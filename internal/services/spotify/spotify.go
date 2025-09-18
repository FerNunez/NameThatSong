package spotify

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/FerNunez/NameThatSong/internal/config"
	"github.com/FerNunez/NameThatSong/internal/pkg/utils"
	"github.com/FerNunez/NameThatSong/internal/repository"
	"github.com/FerNunez/NameThatSong/internal/repository/database"
	"github.com/FerNunez/NameThatSong/internal/services/cache"
	"github.com/redis/go-redis/v9"
)

// Spotify provides a comprehensive Spotify API service
type Spotify struct {
	config     *config.SpotifyConfig
	tokenStore repository.SpotifyTokenStore
	dataStore  repository.SpotifyDataStore
	cache      cache.SpotifyCache
	httpClient *http.Client
}

// NewSpotifyService creates a new comprehensive Spotify service
func NewSpotifyService(
	config *config.SpotifyConfig,
	db *sql.DB,
	redisClient *redis.Client,
) (SpotifyService, error) {
	if config == nil {
		return nil, fmt.Errorf("spotify config is required")
	}

	// Initialize token store
	if db == nil {
		return nil, fmt.Errorf("database connection is required for token store")
	}

	// Create database queries instance and encryptor
	queries := database.New(db)
	encryptor, err := utils.NewTokenEncryptor([]byte(config.EncryptionKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create token encryptor: %v", err)
	}
	tokenStore := repository.NewSQLSpotifyTokenStore(queries, encryptor)

	// Initialize Spotify data store for three-tier caching
	dataStore := repository.NewSQLSpotifyDataStore(queries)

	// Initialize cache (Redis only now)
	var spotifyCache cache.SpotifyCache
	if redisClient != nil {
		spotifyCache = cache.NewRedisSpotifyCache(redisClient)
	} else {
		return nil, fmt.Errorf("redis client is required for caching")
	}

	//TODO: add better client config?
	// Create shared HTTP client with reasonable defaults
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	return &Spotify{
		config:     config,
		tokenStore: tokenStore,
		dataStore:  dataStore,
		cache:      spotifyCache,
		httpClient: httpClient,
	}, nil
}
