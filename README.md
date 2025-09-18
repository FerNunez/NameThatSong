# NameThatSong

**NameThatSong** is a Spotify-integrated web application featuring two main functionalities:

1. **Playlist Enhancement**: Import playlists from Spotify, create and manage local playlists, and export them back to your Spotify account
2. **Music Guessing Game**: Select a playlist and play a song guessing game to test your music knowledge

The application integrates deeply with Spotify's Web API for seamless music management and gameplay, featuring a smart three-tier caching strategy for optimal performance.

## Features

### Playlist Management
- **Import from Spotify**: Fetch user's existing Spotify playlists
- **Local Playlist Management**: Create, edit, and organize playlists locally
- **Export to Spotify**: Push local playlists back to user's Spotify account
- **Sync Management**: Keep local and Spotify playlists synchronized

### Music Guessing Game
- **Playlist Selection**: Choose from imported or local playlists for gameplay
- **Interactive Gameplay**: Real-time song playback with guess validation
- **Score Tracking**: Track accuracy, response time, and game statistics
- **Multiple Difficulties**: Easy, medium, and hard modes with different time limits

## Tech Stack

### Backend
- **Go 1.23+**: Core backend development with Chi router
- **PostgreSQL**: Primary database for playlist and game data
- **Redis**: Caching layer for Spotify API responses and game state
- **SQLC**: Type-safe SQL query generation

### Frontend
- **HTMX**: Dynamic web interactions without JavaScript frameworks
- **Templ**: Type-safe HTML templating in Go
- **TailwindCSS**: Utility-first CSS framework for styling

### External Integrations
- **Spotify Web API**: OAuth 2.0 authentication and playlist management
- **Spotify Web Playback SDK**: Real-time music playback control

### Development Tools
- **Air**: Hot reloading for development
- **Docker Compose**: Local infrastructure setup
- **GitHub Actions**: CI/CD pipeline with automated testing

## Data Strategy

The application implements a **three-tier data retrieval strategy** for optimal performance:

1. **Cache Layer (Redis)**: Fast retrieval of frequently accessed Spotify data
2. **Database Layer (PostgreSQL)**: Persistent storage for user playlists and game data
3. **Spotify API**: Authoritative source when cache/database miss occurs

This approach minimizes API calls, reduces latency, and provides a smooth user experience.

## Acknowledgments

- This project is a fork of [GoTTH](https://github.com/TomDoesTech/GOTTH), a skeleton framework that provided the foundation for development.  
- Thanks to **Spotify** for their fantastic Web API.

