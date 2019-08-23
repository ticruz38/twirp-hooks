package auth

// AuthenticatedUser The user currently authenticated to the app
type AuthenticatedUser struct {
	username        string
	uid             int
	roles           []string
	contractAddress string
	accountAddress  string
	clientID        int
	jwt             string
}
