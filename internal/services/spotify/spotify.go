package spotify

import (
	"database/sql"
	"fmt"

	"github.com/FerNunez/NameThatSong/internal/config"
	"github.com/FerNunez/NameThatSong/internal/pkg/utils"
	"github.com/FerNunez/NameThatSong/internal/repository"
	"github.com/FerNunez/NameThatSong/internal/repository/database"
	"github.com/FerNunez/NameThatSong/internal/services/cache"
	"github.com/redis/go-redis/v9"
)

// SpotifyService provides a comprehensive Spotify API service
type SpotifyService struct {
	config         *config.SpotifyConfig
	tokenStore     repository.SpotifyTokenStore
	cache          cache.SpotifyCache
	searchService  *SpotifySearchService
	fetchService   *SpotifyFetchService
	authService    *SpotifyAuthService
	playerService  *SpotifyPlayerService
}

// NewSpotifyService creates a new comprehensive Spotify service
func NewSpotifyService(
	config *config.SpotifyConfig,
	db *sql.DB,
	redisClient *redis.Client,
) (*SpotifyService, error) {
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

	// Initialize cache (Redis only now)
	var spotifyCache cache.SpotifyCache
	if redisClient != nil {
		spotifyCache = cache.NewRedisSpotifyCache(redisClient)
	} else {
		return nil, fmt.Errorf("Redis client is required for caching")
	}

	// Initialize services
	searchService := NewSpotifySearchService(config, spotifyCache)
	fetchService := NewSpotifyFetchService(config, spotifyCache)
	authService := NewSpotifyAuthService(config, tokenStore, spotifyCache)
	playerService := NewSpotifyPlayerService(config)

	return &SpotifyService{
		config:        config,
		tokenStore:    tokenStore,
		cache:         spotifyCache,
		searchService: searchService,
		fetchService:  fetchService,
		authService:   authService,
		playerService: playerService,
	}, nil
}

// GetAuthService returns the authentication service
func (s *SpotifyService) GetAuthService() *SpotifyAuthService {
	return s.authService
}

// GetSearchService returns the search service
func (s *SpotifyService) GetSearchService() *SpotifySearchService {
	return s.searchService
}

// GetFetchService returns the fetch service
func (s *SpotifyService) GetFetchService() *SpotifyFetchService {
	return s.fetchService
}

// GetPlayerService returns the player service
func (s *SpotifyService) GetPlayerService() *SpotifyPlayerService {
	return s.playerService
}

// GetCache returns the cache service
func (s *SpotifyService) GetCache() cache.SpotifyCache {
	return s.cache
}

// GetConfig returns the Spotify configuration
func (s *SpotifyService) GetConfig() *config.SpotifyConfig {
	return s.config
}

