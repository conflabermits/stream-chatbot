package spotify_client

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

var client *spotify.Client
var ctx = context.Background()

func Initialize(clientID, clientSecret string, token *oauth2.Token) {
	auth := spotifyauth.New(
		spotifyauth.WithClientID(clientID),
		spotifyauth.WithClientSecret(clientSecret),
	)
	client = spotify.New(auth.Client(ctx, token))
}

func formatDuration(ms int) string {
	seconds := ms / 1000
	minutes := seconds / 60
	seconds = seconds % 60
	return fmt.Sprintf("%d:%02d", minutes, seconds)
}

func Search(query string) string {
	results, err := client.Search(ctx, query, spotify.SearchTypeTrack)
	if err != nil {
		log.Println("Error searching Spotify:", err)
		return "Error searching Spotify."
	}
	
	if results.Tracks == nil || len(results.Tracks.Tracks) == 0 {
		return "No results found for your search."
	}

	var msgs []string
	msgs = append(msgs, "Top Results:")
	limit := 5
	if len(results.Tracks.Tracks) < limit {
		limit = len(results.Tracks.Tracks)
	}

	for i := 0; i < limit; i++ {
		t := results.Tracks.Tracks[i]
		artistNames := []string{}
		for _, a := range t.Artists {
			artistNames = append(artistNames, a.Name)
		}
		duration := formatDuration(int(t.Duration))
		msgs = append(msgs, fmt.Sprintf("%d) %s - %s [%s] (ID: %s)", i+1, strings.Join(artistNames, ", "), t.Name, duration, string(t.ID)))
	}
	
	return strings.Join(msgs, " // ")
}

func GetInfo(id string) string {
	trackID := spotify.ID(id)
	t, err := client.GetTrack(ctx, trackID)
	if err != nil {
		log.Println("Error getting track info:", err)
		return "Error getting track info. Ensure the ID is valid."
	}

	artistNames := []string{}
	for _, a := range t.Artists {
		artistNames = append(artistNames, a.Name)
	}

	duration := formatDuration(int(t.Duration))
	releaseDate := t.Album.ReleaseDate
	albumName := t.Album.Name

	return fmt.Sprintf("Info: %s - %s // Album: %s (%s) // Length: %s // ID: %s", strings.Join(artistNames, ", "), t.Name, albumName, releaseDate, duration, string(t.ID))
}

func AddToQueue(queryOrId string) string {
	trackID := spotify.ID(queryOrId)
	
	// Check if it's an ID or a query. A valid ID is 22 chars of base62
	// If getting track fails, try searching.
	t, err := client.GetTrack(ctx, trackID)
	if err != nil || t == nil {
		// Try search
		results, err := client.Search(ctx, queryOrId, spotify.SearchTypeTrack)
		if err != nil || results.Tracks == nil || len(results.Tracks.Tracks) == 0 {
			return "Could not find a song matching that query or ID."
		}
		t = &results.Tracks.Tracks[0]
		trackID = t.ID
	}

	err = client.QueueSong(ctx, trackID)
	if err != nil {
		log.Println("Error adding to queue:", err)
		return "Error adding song to the queue."
	}

	artistNames := []string{}
	for _, a := range t.Artists {
		artistNames = append(artistNames, a.Name)
	}

	return fmt.Sprintf("Added to queue: %s - %s", strings.Join(artistNames, ", "), t.Name)
}

func GetQueue() string {
	queue, err := client.GetQueue(ctx)
	if err != nil {
		log.Println("Error getting queue:", err)
		return "Error retrieving the queue."
	}

	if len(queue.Items) == 0 {
		return "The queue is currently empty."
	}

	var msgs []string
	msgs = append(msgs, "Upcoming:")
	limit := 10
	if len(queue.Items) < limit {
		limit = len(queue.Items)
	}

	for i := 0; i < limit; i++ {
		t := queue.Items[i]
		artistNames := []string{}
		for _, a := range t.Artists {
			artistNames = append(artistNames, a.Name)
		}
		msgs = append(msgs, fmt.Sprintf("%d) %s - %s (ID: %s)", i+1, strings.Join(artistNames, ", "), t.Name, string(t.ID)))
	}

	return strings.Join(msgs, " // ")
}

func Pause() string {
	err := client.Pause(ctx)
	if err != nil {
		return "Failed to pause."
	}
	return "Playback paused."
}

func Play() string {
	err := client.Play(ctx)
	if err != nil {
		return "Failed to play."
	}
	return "Playback resumed."
}

func Next() string {
	err := client.Next(ctx)
	if err != nil {
		return "Failed to skip to next track."
	}
	return "Skipped to next track."
}
