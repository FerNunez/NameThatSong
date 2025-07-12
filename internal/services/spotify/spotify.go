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

// SpotifyService provides a comprehensive Spotify API service
type SpotifyService struct {
	searchService *SpotifySearchService
	fetchService  *SpotifyFetchService
	authService   *SpotifyAuthService
	playerService *SpotifyPlayerService
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

	//TODO: add better client config?
	// Create shared HTTP client with reasonable defaults
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Initialize services with shared HTTP client
	authService := NewSpotifyAuthService(config, tokenStore, spotifyCache, httpClient)
	searchService := NewSpotifySearchService(config, spotifyCache, authService, httpClient)
	fetchService := NewSpotifyFetchService(config, spotifyCache, authService, httpClient)
	playerService := NewSpotifyPlayerService(config, authService, httpClient)

	return &SpotifyService{
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
