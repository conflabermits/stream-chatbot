package player

import (
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Track represents a single music request in the queue.
type Track struct {
	ID          string `json:"id"`           // Video ID
	Title       string `json:"title"`        // Video Title
	Artist      string `json:"artist"`       // Channel or Artist
	Duration    int    `json:"duration"`     // Length in seconds
	RequestedBy string `json:"requested_by"` // Twitch Username
	Thumbnail   string `json:"thumbnail"`    // Thumbnail URL (optional)
}

var (
	queue        []Track
	currentTrack *Track
	queueMutex   sync.Mutex
	IsPaused     bool

	// Rate limiting maps
	userRequests   = make(map[string][]time.Time)
	rateLimitMutex sync.Mutex

	// Configuration
	MaxDurationSeconds = 1200 // 20 minutes
	CooldownDuration   = 10 * time.Minute
	RateLimitEnabled   = true
)

// ToggleRateLimit toggles the rate limiting feature.
func ToggleRateLimit() bool {
	rateLimitMutex.Lock()
	defer rateLimitMutex.Unlock()
	RateLimitEnabled = !RateLimitEnabled
	return RateLimitEnabled
}

// AddTrack adds a track to the queue, applying duration and rate limit checks.
func AddTrack(track Track, isMod bool) error {
	if !isMod {
		if track.Duration > MaxDurationSeconds {
			return errors.New("track exceeds maximum allowed duration")
		}

		rateLimitMutex.Lock()
		if RateLimitEnabled {
			now := time.Now()
			var recent []time.Time
			for _, t := range userRequests[track.RequestedBy] {
				if now.Sub(t) < CooldownDuration {
					recent = append(recent, t)
				}
			}
			userRequests[track.RequestedBy] = recent

			if len(recent) >= 2 {
				rateLimitMutex.Unlock()
				return errors.New("you are on cooldown, please wait before requesting again")
			}
			userRequests[track.RequestedBy] = append(recent, now)
		}
		rateLimitMutex.Unlock()
	}

	queueMutex.Lock()
	defer queueMutex.Unlock()

	queue = append(queue, track)

	return nil
}

// PopNext moves the first item in the queue to currentTrack.
func PopNext() *Track {
	queueMutex.Lock()
	defer queueMutex.Unlock()
	IsPaused = false
	return popNextInternal()
}

// popNextInternal is a helper function that must be called with queueMutex held.
func popNextInternal() *Track {
	if len(queue) == 0 {
		currentTrack = nil
		return nil
	}

	next := queue[0]
	queue = queue[1:]
	currentTrack = &next
	return currentTrack
}

// ClearCurrent stops the currently playing track and leaves the player waiting.
func ClearCurrent() {
	queueMutex.Lock()
	defer queueMutex.Unlock()
	currentTrack = nil
	IsPaused = false
}

// PlayOrResume resumes playback if paused, or pops next track if stopped.
func PlayOrResume() error {
	queueMutex.Lock()
	defer queueMutex.Unlock()

	if currentTrack != nil {
		if IsPaused {
			IsPaused = false
			return nil
		}
		return errors.New("track is already playing")
	}

	if len(queue) == 0 {
		return errors.New("queue is empty")
	}

	IsPaused = false
	popNextInternal()
	return nil
}

// Pause pauses the current track.
func Pause() error {
	queueMutex.Lock()
	defer queueMutex.Unlock()

	if currentTrack == nil {
		return errors.New("nothing is playing")
	}

	if IsPaused {
		return errors.New("already paused")
	}

	IsPaused = true
	return nil
}

// GetIsPaused returns whether the player is paused
func GetIsPaused() bool {
	queueMutex.Lock()
	defer queueMutex.Unlock()
	return IsPaused
}

// SkipCurrent immediately stops the current track.
func SkipCurrent() {
	ClearCurrent()
}

// GetQueue returns a copy of the current queue.
func GetQueue() []Track {
	queueMutex.Lock()
	defer queueMutex.Unlock()

	qCopy := make([]Track, len(queue))
	copy(qCopy, queue)
	return qCopy
}

// GetCurrentTrack returns the currently playing track.
func GetCurrentTrack() *Track {
	queueMutex.Lock()
	defer queueMutex.Unlock()
	return currentTrack
}

// RemoveTrack removes a track based on the smart removal logic.
func RemoveTrack(query string, isMod bool, requester string) (string, error) {
	queueMutex.Lock()
	defer queueMutex.Unlock()

	if len(queue) == 0 {
		return "", errors.New("the queue is empty")
	}

	if query == "" {
		// Remove most recently added
		for i := len(queue) - 1; i >= 0; i-- {
			if isMod || queue[i].RequestedBy == requester {
				track := queue[i]
				queue = append(queue[:i], queue[i+1:]...)
				return track.Title, nil
			}
		}
		return "", errors.New("no tracks found to remove")
	}

	// Try removing by index
	if idx, err := strconv.Atoi(query); err == nil {
		idx = idx - 1 // 1-indexed
		if idx >= 0 && idx < len(queue) {
			if isMod || queue[idx].RequestedBy == requester {
				track := queue[idx]
				queue = append(queue[:idx], queue[idx+1:]...)
				return track.Title, nil
			} else {
				return "", errors.New("you do not have permission to remove this track")
			}
		}
	}

	// String match
	queryLower := strings.ToLower(query)
	for i, track := range queue {
		if strings.Contains(strings.ToLower(track.Title), queryLower) || strings.Contains(strings.ToLower(track.Artist), queryLower) {
			if isMod || track.RequestedBy == requester {
				trackTitle := track.Title
				queue = append(queue[:i], queue[i+1:]...)
				return trackTitle, nil
			}
		}
	}

	return "", errors.New("no matching track found or you lack permissions to remove it")
}
