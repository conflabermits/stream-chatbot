package player

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

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

// SearchTrack determines the source and fetches the necessary metadata.
func SearchTrack(query string, requestedBy string) (*Track, error) {
	// Simple source detection
	if strings.Contains(query, "bandcamp.com") {
		return fetchBandcampMetadata(query, requestedBy)
	}

	// Default to YouTube
	return searchYouTube(query, requestedBy)
}

func searchYouTube(query string, requestedBy string) (*Track, error) {
	err := initYoutube()
	if err != nil {
		return nil, fmt.Errorf("youtube api not initialized: %v", err)
	}

	videoID := ""
	// Check if query is a youtube URL
	if strings.Contains(query, "youtube.com") || strings.Contains(query, "youtu.be") {
		u, err := url.Parse(query)
		if err == nil {
			if strings.Contains(query, "youtu.be") {
				videoID = strings.TrimPrefix(u.Path, "/")
			} else {
				videoID = u.Query().Get("v")
			}
		}
	}

	if videoID == "" {
		// Perform a search
		call := ytService.Search.List([]string{"id", "snippet"}).Q(query).MaxResults(1).Type("video")
		response, err := call.Do()
		if err != nil {
			return nil, fmt.Errorf("error searching youtube: %v", err)
		}
		if len(response.Items) == 0 {
			return nil, errors.New("no youtube results found")
		}
		videoID = response.Items[0].Id.VideoId
	}

	// Fetch duration and full details
	videosCall := ytService.Videos.List([]string{"snippet", "contentDetails"}).Id(videoID)
	videosResponse, err := videosCall.Do()
	if err != nil {
		return nil, fmt.Errorf("error fetching video details: %v", err)
	}
	if len(videosResponse.Items) == 0 {
		return nil, errors.New("youtube video not found")
	}

	item := videosResponse.Items[0]
	durationStr := item.ContentDetails.Duration
	durationSeconds := parseISO8601Duration(durationStr)

	thumbnail := ""
	if item.Snippet.Thumbnails != nil && item.Snippet.Thumbnails.Default != nil {
		thumbnail = item.Snippet.Thumbnails.Default.Url
	}

	return &Track{
		ID:          videoID,
		Title:       item.Snippet.Title,
		Artist:      item.Snippet.ChannelTitle,
		Source:      "youtube",
		Duration:    durationSeconds,
		RequestedBy: requestedBy,
		Thumbnail:   thumbnail,
	}, nil
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

func fetchBandcampMetadata(pageURL string, requestedBy string) (*Track, error) {
	// Ensure URL has http scheme
	if !strings.HasPrefix(pageURL, "http") {
		pageURL = "https://" + pageURL
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(pageURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch bandcamp page: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bandcamp returned status %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	html := string(bodyBytes)

	// Extract Title
	title := "Unknown Bandcamp Track"
	titleRe := regexp.MustCompile(`<meta property="og:title" content="([^"]+)">`)
	if m := titleRe.FindStringSubmatch(html); len(m) > 1 {
		title = m[1]
	}

	// Extract Artist (usually from the site name or part of the title)
	artist := "Bandcamp Artist"
	artistRe := regexp.MustCompile(`<meta property="og:site_name" content="([^"]+)">`)
	if m := artistRe.FindStringSubmatch(html); len(m) > 1 {
		artist = m[1]
	}

	// Extract Embed URL to get Album/Track IDs
	embedURL := ""
	embedRe := regexp.MustCompile(`<meta property="og:video" content="([^"]+)">`)
	if m := embedRe.FindStringSubmatch(html); len(m) > 1 {
		embedURL = m[1]
	} else {
		// Fallback for getting embed URL params from trkinfo
		albumRe := regexp.MustCompile(`"album_id":(\d+)`)
		trackRe := regexp.MustCompile(`"track_id":(\d+)`)
		albumMatch := albumRe.FindStringSubmatch(html)
		trackMatch := trackRe.FindStringSubmatch(html)
		
		if len(albumMatch) > 1 && len(trackMatch) > 1 {
			embedURL = fmt.Sprintf("https://bandcamp.com/EmbeddedPlayer/album=%s/track=%s/size=small/transparent=true/", albumMatch[1], trackMatch[1])
		} else {
			return nil, errors.New("could not find bandcamp track or album id")
		}
	}

	// Make sure embed uses a standard simple layout
	if strings.Contains(embedURL, "size=large") {
		embedURL = strings.ReplaceAll(embedURL, "size=large", "size=small")
	}
	if !strings.Contains(embedURL, "transparent=true") {
		embedURL += "transparent=true/"
	}

	// Extract Duration (TrkInfo or ItemData)
	duration := 0
	durationRe := regexp.MustCompile(`"duration":([0-9.]+)`)
	if m := durationRe.FindStringSubmatch(html); len(m) > 1 {
		d, err := strconv.ParseFloat(m[1], 64)
		if err == nil {
			duration = int(d)
		}
	}

	if duration == 0 {
		// Look for video:duration meta
		metaDurRe := regexp.MustCompile(`<meta property="video:duration" content="(\d+)">`)
		if m := metaDurRe.FindStringSubmatch(html); len(m) > 1 {
			if d, err := strconv.Atoi(m[1]); err == nil {
				duration = d
			}
		}
	}

	if duration == 0 {
		return nil, errors.New("could not determine track duration from bandcamp page")
	}

	return &Track{
		ID:          pageURL,
		Title:       title,
		Artist:      artist,
		Source:      "bandcamp",
		Duration:    duration,
		RequestedBy: requestedBy,
		BandcampURL: embedURL,
	}, nil
}
