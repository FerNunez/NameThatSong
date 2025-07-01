package models

// / Spotify Search
type TrackSearch struct {
	Name       string
	Id         string
	Popularity int
	ArtistList []string
}

type AlbumSearch struct {
	Name       string
	Id         string
	ImageUrl   string
	ArtistList []string
}
type ArtistSearch struct {
	Name       string
	Id         string
	ImageUrl   string
	Popularity int
}

type PlaylistSearch struct {
	Name     string
	Id       string
	ImageUrl string
}

// / Spotify Data
type TrackData struct {
	DiscNumber  int
	DurationMs  int
	ID          string
	Name        string
	TrackNumber int
	Popularity  int
	Explicit    bool
}

type AlbumData struct {
	AlbumType   string
	TotalTracks int
	ID          string
	ImagesURL   string
	Name        string
	ReleaseDate string
}

type ArtistData struct {
	Id         string
	Name       string
	ImageUrl   string
	Popularity int
}

type PlaylistData struct {
	Description    string
	FollowersTotal int
	ID             string
	ImageUrl       string
	Name           string
	Public         bool
	TotalTracks    int
}
