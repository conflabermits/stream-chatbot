package auth

import (
	"fmt"
	"log"
	"net/http"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

var (
	spotifyTokenChan chan *oauth2.Token
	spotifyAuthApp   *spotifyauth.Authenticator
	spotifyState     = "spoopify-state-string" // In production, this should be random
)

func SpotifyAuth(TokenChan chan *oauth2.Token, clientID string, clientSecret string) {
	spotifyTokenChan = TokenChan

	// NOTE: The Spotify Web API requires users to have an active Spotify Premium
	// subscription to access playback features (like adding to queue, playing, pausing).
	// NOTE: Spotify API blocks "http://localhost" as a valid redirect URI and typically
	// requires HTTPS. However, it explicitly permits HTTP for loopback IPs like "http://127.0.0.1"
	spotifyAuthApp = spotifyauth.New(
		spotifyauth.WithRedirectURL("http://127.0.0.1:8081/callback"),
		spotifyauth.WithScopes(
			spotifyauth.ScopeUserReadPlaybackState,
			spotifyauth.ScopeUserModifyPlaybackState,
			spotifyauth.ScopeUserReadCurrentlyPlaying,
		),
		spotifyauth.WithClientID(clientID),
		spotifyauth.WithClientSecret(clientSecret),
	)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", completeSpotifyAuth)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		url := spotifyAuthApp.AuthURL(spotifyState)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	})

	log.Println("Started running Spotify auth on http://127.0.0.1:8081/")
	url := spotifyAuthApp.AuthURL(spotifyState)
	fmt.Println("Please log in to Spotify by visiting the following page in your browser:", url)
	log.Println(http.ListenAndServe(":8081", mux))
}

func completeSpotifyAuth(w http.ResponseWriter, r *http.Request) {
	tok, err := spotifyAuthApp.Token(r.Context(), spotifyState, r)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		log.Println("Error getting Spotify token:", err)
		return
	}
	if st := r.FormValue("state"); st != spotifyState {
		http.NotFound(w, r)
		log.Printf("State mismatch: %s != %s\n", st, spotifyState)
		return
	}

	client := spotify.New(spotifyAuthApp.Client(r.Context(), tok))
	user, err := client.CurrentUser(r.Context())
	if err != nil {
		log.Println("Error getting current user:", err)
	} else {
		log.Printf("Logged in to Spotify as: %s\n", user.ID)
	}

	go func() {
		spotifyTokenChan <- tok
	}()

	fmt.Fprintf(w, "Successfully logged in to Spotify! You can close this window.")
}
