package player

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"stream-chatbot/common"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

var ytService *youtube.Service

func initYoutube() error {
	if ytService != nil {
		return nil
	}
	apiKey := common.ChatbotCreds["YouTubeAPIKey"]
	if apiKey == "" {
		return errors.New("YouTubeAPIKey is not set in credentials")
	}

	ctx := context.Background()
	service, err := youtube.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return err
	}
	ytService = service
	return nil
}

// SearchTrack searches YouTube for the given query and returns matching tracks.
func SearchTrack(query string, requestedBy string) ([]Track, error) {
	return searchYouTube(query, requestedBy)
}

func searchYouTube(query string, requestedBy string) ([]Track, error) {
	err := initYoutube()
	if err != nil {
		return nil, fmt.Errorf("youtube api not initialized: %v", err)
	}

	var videoIDs []string
	log.Printf("[YouTube API] Searching for query: %s\n", query)

	// Check if query is a youtube URL
	if strings.Contains(query, "youtube.com") || strings.Contains(query, "youtu.be") {
		u, err := url.Parse(query)
		if err == nil {
			if strings.Contains(query, "youtu.be") {
				videoIDs = append(videoIDs, strings.TrimPrefix(u.Path, "/"))
			} else {
				videoIDs = append(videoIDs, u.Query().Get("v"))
			}
		}
	}

	if len(videoIDs) == 0 || videoIDs[0] == "" {
		// Perform a search
		call := ytService.Search.List([]string{"id", "snippet"}).Q(query).MaxResults(3).Type("video")
		response, err := call.Do()
		if err != nil {
			log.Printf("[YouTube API] Search error: %v\n", err)
			return nil, fmt.Errorf("error searching youtube: %v", err)
		}
		if len(response.Items) == 0 {
			log.Printf("[YouTube API] No results found for query: %s\n", query)
			return nil, errors.New("no youtube results found")
		}
		for _, item := range response.Items {
			videoIDs = append(videoIDs, item.Id.VideoId)
		}
	}

	// Fetch duration and full details
	videosCall := ytService.Videos.List([]string{"snippet", "contentDetails"}).Id(strings.Join(videoIDs, ","))
	videosResponse, err := videosCall.Do()
	if err != nil {
		log.Printf("[YouTube API] Videos List error: %v\n", err)
		return nil, fmt.Errorf("error fetching video details: %v", err)
	}
	if len(videosResponse.Items) == 0 {
		log.Printf("[YouTube API] Videos not found for IDs: %v\n", videoIDs)
		return nil, errors.New("youtube video not found")
	}

	var tracks []Track
	for _, item := range videosResponse.Items {
		durationStr := item.ContentDetails.Duration
		durationSeconds := parseISO8601Duration(durationStr)

		thumbnail := ""
		if item.Snippet.Thumbnails != nil && item.Snippet.Thumbnails.Default != nil {
			thumbnail = item.Snippet.Thumbnails.Default.Url
		}

		tracks = append(tracks, Track{
			ID:          item.Id,
			Title:       item.Snippet.Title,
			Artist:      item.Snippet.ChannelTitle,
			Duration:    durationSeconds,
			RequestedBy: requestedBy,
			Thumbnail:   thumbnail,
		})
	}

	return tracks, nil
}

func parseISO8601Duration(duration string) int {
	// ISO 8601 duration format: PT#H#M#S
	re := regexp.MustCompile(`PT(?:(\d+)H)?(?:(\d+)M)?(?:(\d+)S)?`)
	matches := re.FindStringSubmatch(duration)
	if len(matches) == 0 {
		return 0
	}

	hours, _ := strconv.Atoi(matches[1])
	minutes, _ := strconv.Atoi(matches[2])
	seconds, _ := strconv.Atoi(matches[3])

	return hours*3600 + minutes*60 + seconds
}
