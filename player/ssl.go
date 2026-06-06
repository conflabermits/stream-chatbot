package player

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"stream-chatbot/common"
)

type SSLQueueResponse struct {
	List []SSLQueueItem `json:"list"`
}

type SSLQueueItem struct {
	Song struct {
		Title   string `json:"title"`
		Artist  string `json:"artist"`
		Comment string `json:"comment"`
	} `json:"song"`
	Requests []struct {
		Name string `json:"name"`
	} `json:"requests"`
}

// FetchNextFromSSL gets the top song from Streamer Songlist queue.
func FetchNextFromSSL() (*Track, error) {
	channel := common.ChatbotCreds["TwitchChannel"]
	channel = strings.TrimPrefix(channel, "#")
	if channel == "" {
		channel = "conflabermits"
	}

	url := fmt.Sprintf("https://api.streamersonglist.com/v1/streamers/%s/queue", channel)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch SSL queue: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var qResp SSLQueueResponse
	err = json.Unmarshal(body, &qResp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SSL queue: %v", err)
	}

	if len(qResp.List) == 0 {
		return nil, errors.New("queue is empty")
	}

	top := qResp.List[0]

	requestedBy := "Unknown"
	if len(top.Requests) > 0 {
		requestedBy = top.Requests[0].Name
	}

	// Determine YouTube Video ID
	videoID := extractVideoID(top.Song.Comment)

	var tracks []Track
	if videoID != "" {
		// Get full details from YouTube API (duration, etc)
		tracks, err = SearchTrack(videoID, requestedBy)
	} else {
		// Fallback: Search YouTube using Artist - Title
		query := fmt.Sprintf("%s - %s", top.Song.Artist, top.Song.Title)
		tracks, err = SearchTrack(query, requestedBy)
	}

	if err != nil || len(tracks) == 0 {
		return nil, errors.New("could not find youtube video for track")
	}

	track := tracks[0]
	// Override with SSL title/artist for overlay consistency
	track.Title = top.Song.Title
	track.Artist = top.Song.Artist

	return &track, nil
}

func extractVideoID(comment string) string {
	if strings.Contains(comment, "v=") {
		parts := strings.Split(comment, "v=")
		if len(parts) > 1 {
			id := strings.Split(parts[1], "&")[0]
			return id
		}
	} else if strings.Contains(comment, "youtu.be/") {
		parts := strings.Split(comment, "youtu.be/")
		if len(parts) > 1 {
			id := strings.Split(parts[1], "?")[0]
			return id
		}
	}
	return ""
}
