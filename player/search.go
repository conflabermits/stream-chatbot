package player

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"log"
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
func SearchTrack(query string, requestedBy string, preferredSource string) ([]Track, error) {
	// Simple source detection
	if strings.Contains(query, "bandcamp.com") {
		return fetchBandcampMetadata(query, requestedBy)
	}

	if strings.Contains(query, "youtube.com") || strings.Contains(query, "youtu.be") {
		return searchYouTube(query, requestedBy)
	}

	if preferredSource == "bandcamp" {
		return searchBandcamp(query, requestedBy)
	}

	// Default to YouTube
	return searchYouTube(query, requestedBy)
}

type bcSearchPayload struct {
	SearchText   string `json:"search_text"`
	SearchFilter string `json:"search_filter"`
	FullPage     bool   `json:"full_page"`
}

type bcSearchResponse struct {
	Auto struct {
		Results []struct {
			Type        string `json:"type"`
			ItemUrlPath string `json:"item_url_path"`
		} `json:"results"`
	} `json:"auto"`
}

func searchBandcamp(query string, requestedBy string) ([]Track, error) {
	apiURL := "https://bandcamp.com/api/bcsearch_public_api/1/autocomplete_elastic"
	log.Printf("[Bandcamp API] Searching for query: %s\n", query)
	
	payload := bcSearchPayload{
		SearchText:   query,
		SearchFilter: "t", // 't' for tracks
		FullPage:     false,
	}
	
	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[Bandcamp API] Marshal error: %v\n", err)
		return nil, fmt.Errorf("failed to marshal bandcamp payload: %v", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("[Bandcamp API] Request creation error: %v\n", err)
		return nil, fmt.Errorf("failed to create bandcamp request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[Bandcamp API] HTTP error: %v\n", err)
		return nil, fmt.Errorf("failed to fetch bandcamp search: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("[Bandcamp API] Bad status code: %d\n", resp.StatusCode)
		return nil, fmt.Errorf("bandcamp api returned status %d", resp.StatusCode)
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	
	var searchResp bcSearchResponse
	if err := json.Unmarshal(bodyBytes, &searchResp); err != nil {
		log.Printf("[Bandcamp API] Unmarshal error: %v\n", err)
		return nil, fmt.Errorf("failed to decode bandcamp response: %v", err)
	}

	var tracks []Track
	seen := make(map[string]bool)
	for _, result := range searchResp.Auto.Results {
		if len(tracks) >= 3 {
			break
		}
		if result.Type == "t" && result.ItemUrlPath != "" {
			trackURL := result.ItemUrlPath
			if !seen[trackURL] {
				seen[trackURL] = true
				trackList, err := fetchBandcampMetadata(trackURL, requestedBy)
				if err == nil && len(trackList) > 0 {
					tracks = append(tracks, trackList[0])
				}
			}
		}
	}

	if len(tracks) == 0 {
		log.Printf("[Bandcamp API] No tracks found for query: %s\n", query)
		return nil, errors.New("no bandcamp results found")
	}
	return tracks, nil
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
			Source:      "youtube",
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

func fetchBandcampMetadata(pageURL string, requestedBy string) ([]Track, error) {
	// Ensure URL has http scheme
	if !strings.HasPrefix(pageURL, "http") {
		pageURL = "https://" + pageURL
	}
	log.Printf("[Bandcamp Scrape] Fetching URL: %s\n", pageURL)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(pageURL)
	if err != nil {
		log.Printf("[Bandcamp Scrape] HTTP error: %v\n", err)
		return nil, fmt.Errorf("failed to fetch bandcamp page: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("[Bandcamp Scrape] Bad status code: %d\n", resp.StatusCode)
		return nil, fmt.Errorf("bandcamp returned status %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[Bandcamp Scrape] ReadBody error: %v\n", err)
		return nil, err
	}
	pageHTML := string(bodyBytes)

	// Extract Title
	title := "Unknown Bandcamp Track"
	titleRe := regexp.MustCompile(`<meta property="og:title" content="([^"]+)">`)
	if m := titleRe.FindStringSubmatch(pageHTML); len(m) > 1 {
		title = m[1]
	}

	// Extract Artist
	artist := "Bandcamp Artist"
	artistRe := regexp.MustCompile(`<meta property="og:site_name" content="([^"]+)">`)
	if m := artistRe.FindStringSubmatch(pageHTML); len(m) > 1 {
		artist = m[1]
	}

	embedURL := ""
	duration := 0

	// Use data-tralbum robust extraction
	tralbumRe := regexp.MustCompile(`data-tralbum="([^"]+)"`)
	if m := tralbumRe.FindStringSubmatch(pageHTML); len(m) > 1 {
		decodedJSON := html.UnescapeString(m[1])
		
		albumRe := regexp.MustCompile(`"album_id":\s*(\d+)`)
		trackRe := regexp.MustCompile(`"track_id":\s*(\d+)`)
		durationRe := regexp.MustCompile(`"duration":\s*([0-9.]+)`)
		
		albumMatch := albumRe.FindStringSubmatch(decodedJSON)
		trackMatch := trackRe.FindStringSubmatch(decodedJSON)
		durationMatch := durationRe.FindStringSubmatch(decodedJSON)

		if len(albumMatch) > 1 && len(trackMatch) > 1 {
			embedURL = fmt.Sprintf("https://bandcamp.com/EmbeddedPlayer/album=%s/track=%s/size=small/transparent=true/", albumMatch[1], trackMatch[1])
		}
		
		if len(durationMatch) > 1 {
			d, err := strconv.ParseFloat(durationMatch[1], 64)
			if err == nil {
				duration = int(d)
			}
		}
	}

	if embedURL == "" {
		log.Printf("[Bandcamp Scrape] Could not find track/album id for URL: %s\n", pageURL)
		return nil, errors.New("could not find bandcamp track or album id")
	}

	if duration == 0 {
		log.Printf("[Bandcamp Scrape] Could not find duration for URL: %s\n", pageURL)
		return nil, errors.New("could not determine track duration from bandcamp page")
	}

	log.Printf("[Bandcamp Scrape] Success: %s - %s (%d sec)\n", artist, title, duration)

	return []Track{{
		ID:          pageURL,
		Title:       title,
		Artist:      artist,
		Source:      "bandcamp",
		Duration:    duration,
		RequestedBy: requestedBy,
		BandcampURL: embedURL,
	}}, nil
}
