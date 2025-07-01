package spotify

// Token Manager should be a Map/Repository userId to Tokens?
type TokenManager struct {
	State        string
	AccessToken  string
	RefreshToken string
}
