package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Configuration struct
type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// User struct to store Google user info
type User struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

var (
	googleOauthConfig *oauth2.Config
	// Store to save state tokens
	stateTokens = make(map[string]time.Time)
)

func main() {
	// Load configuration from environment variables
	config := Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("REDIRECT_URL"),
	}

	if config.ClientID == "" || config.ClientSecret == "" || config.RedirectURL == "" {
		log.Fatal("Missing required environment variables. Please set GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET, and REDIRECT_URL")
	}

	// Initialize the OAuth2 config
	googleOauthConfig = &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	// Clean expired state tokens periodically
	go cleanExpiredStateTokens()

	// Set up HTTP routes
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/login", handleLogin)
	http.HandleFunc("/identity/google/callback", handleGoogleCallback)
	http.HandleFunc("/profile", handleProfile)
	http.HandleFunc("/logout", handleLogout)

	// Start the server
	fmt.Println("Server starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Handler for home page
func handleHome(w http.ResponseWriter, r *http.Request) {
	html := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Google OAuth Login</title>
		<style>
			body { font-family: Arial, sans-serif; margin: 40px; text-align: center; }
			.login-btn { 
				display: inline-block;
				background-color: #4285F4;
				color: white;
				padding: 10px 20px;
				text-decoration: none;
				border-radius: 5px;
				margin-top: 20px;
			}
		</style>
	</head>
	<body>
		<h1>Welcome to OAuth Login Demo</h1>
		<p>Click the button below to log in with Google</p>
		<a href="/login" class="login-btn">Login with Google</a>
	</body>
	</html>
	`
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// Handler for login route that redirects to Google
func handleLogin(w http.ResponseWriter, r *http.Request) {
	slog.Info("Google Login", "url", r.URL)
	// Generate random state token
	stateToken, err := generateStateToken()
	if err != nil {
		http.Error(w, "Error generating state token", http.StatusInternalServerError)
		return
	}

	// Store the state token with an expiration time (10 minutes)
	stateTokens[stateToken] = time.Now().Add(10 * time.Minute)

	// Redirect to Google's OAuth 2.0 server
	url := googleOauthConfig.AuthCodeURL(stateToken)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// Handler for Google's callback
func handleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	slog.Info("Google Callback", "url", r.URL)
	// Get the state and code from URL parameters
	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")

	// Verify state token
	_, valid := stateTokens[state]
	if !valid {
		http.Error(w, "Invalid state token", http.StatusBadRequest)
		return
	}
	delete(stateTokens, state) // Remove used token

	// Exchange the authorization code for a token
	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get user info
	user, err := getUserInfo(token.AccessToken)
	if err != nil {
		http.Error(w, "Failed to get user info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Create session (in a real app, you would use a proper session management library)
	// For this example, we'll just store user info in a cookie
	sessionToken, err := generateStateToken()
	if err != nil {
		http.Error(w, "Failed to generate session token", http.StatusInternalServerError)
		return
	}

	// Save user info in session (in a real app, this would be stored securely server-side)
	userJSON, err := json.Marshal(user)
	if err != nil {
		http.Error(w, "Failed to encode user data", http.StatusInternalServerError)
		return
	}

	// Set cookies
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "user_data",
		Value:    base64.StdEncoding.EncodeToString(userJSON),
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	})

	// Redirect to profile page
	http.Redirect(w, r, "/profile", http.StatusTemporaryRedirect)
}

// Handler for profile page
func handleProfile(w http.ResponseWriter, r *http.Request) {
	// Check if user is logged in
	_, err := r.Cookie("session_token")
	if err != nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	userDataCookie, err := r.Cookie("user_data")
	if err != nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Decode user data
	userDataJSON, err := base64.StdEncoding.DecodeString(userDataCookie.Value)
	if err != nil {
		http.Error(w, "Invalid user data", http.StatusBadRequest)
		return
	}

	var user User
	if err := json.Unmarshal(userDataJSON, &user); err != nil {
		http.Error(w, "Failed to parse user data", http.StatusInternalServerError)
		return
	}

	slog.Info("Profile User info", "user", user)

	// get base64 of user picture url
	url := user.Picture
	slog.Info("Profile User Picture", "url", url)

	// get user picture
	resp, err := http.Get(url)
	if err != nil {
		http.Error(w, "Failed to get user picture", http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()

	// get user picture
	picture, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read user picture", http.StatusInternalServerError)
		return
	}

	// set user picture
	user.Picture = base64.StdEncoding.EncodeToString(picture)

	user.Picture = "data:image/png;base64," + user.Picture

	// Display profile page
	html := fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<head>
		<title>Profile</title>
		<style>
			body { font-family: Arial, sans-serif; margin: 40px; text-align: center; }
			.profile-container {
				max-width: 600px;
				margin: 0 auto;
				padding: 20px;
				border: 1px solid #ddd;
				border-radius: 5px;
			}
			.user-image {
				width: 100px;
				height: 100px;
				border-radius: 50%%;
				margin-bottom: 20px;
			}
			.logout-btn {
				display: inline-block;
				background-color: #f44336;
				color: white;
				padding: 10px 20px;
				text-decoration: none;
				border-radius: 5px;
				margin-top: 20px;
			}
		</style>
	</head>
	<body>
		<div class="profile-container">
			<h1>Welcome, %s!</h1>
			<img src="%s" alt="Profile picture" class="user-image">
			<p><strong>Email:</strong> %s</p>
			<p><strong>Name:</strong> %s %s</p>
			<a href="/logout" class="logout-btn">Logout</a>
		</div>
	</body>
	</html>
	`, user.Name, user.Picture, user.Email, user.GivenName, user.FamilyName)

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// Handler for logout
func handleLogout(w http.ResponseWriter, r *http.Request) {
	// Clear cookies
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "user_data",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	// Redirect to home page
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

// Helper function to get user info from Google API
func getUserInfo(accessToken string) (*User, error) {
	// Make request to Google's UserInfo API
	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + accessToken)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("getUserInfo: failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var user User
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, err
	}

	slog.Info("User info", "user", user)

	return &user, nil
}

// Helper function to generate a random state token
func generateStateToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

// Function to clean expired state tokens
func cleanExpiredStateTokens() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		for token, expiry := range stateTokens {
			if now.After(expiry) {
				delete(stateTokens, token)
			}
		}
	}
}
