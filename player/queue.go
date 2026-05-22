package player

import (
	"errors"
	"sync"
	"time"
)

// Track represents a single music request in the queue.
type Track struct {
	ID          string `json:"id"`           // Video ID or Bandcamp track URL/ID
	Title       string `json:"title"`        // Video/Track Title
	Artist      string `json:"artist"`       // Channel or Artist
	Source      string `json:"source"`       // "youtube" | "bandcamp"
	Duration    int    `json:"duration"`     // Length in seconds
	RequestedBy string `json:"requested_by"` // Twitch Username
	Thumbnail   string `json:"thumbnail"`    // Thumbnail URL (optional)
	BandcampURL string `json:"bandcamp_url"` // Bandcamp iframe SRC or track URL
}

var (
	queue        []Track
	currentTrack *Track
	queueMutex   sync.Mutex

	// Rate limiting maps
	userLastRequest = make(map[string]time.Time)
	rateLimitMutex  sync.Mutex

	// Configuration
	MaxDurationSeconds = 1200 // 20 minutes
	CooldownDuration   = 10 * time.Minute
)

// AddTrack adds a track to the queue, applying duration and rate limit checks.
func AddTrack(track Track) error {
	if track.Duration > MaxDurationSeconds {
		return errors.New("track exceeds maximum allowed duration")
	}

	rateLimitMutex.Lock()
	lastReq, exists := userLastRequest[track.RequestedBy]
	if exists && time.Since(lastReq) < CooldownDuration {
		rateLimitMutex.Unlock()
		return errors.New("you are on cooldown, please wait before requesting again")
	}
	userLastRequest[track.RequestedBy] = time.Now()
	rateLimitMutex.Unlock()

	queueMutex.Lock()
	defer queueMutex.Unlock()

	queue = append(queue, track)
	
	// If nothing is playing, pop the next track immediately
	if currentTrack == nil {
		popNextInternal()
	}

	return nil
}

// PopNext moves the first item in the queue to currentTrack.
func PopNext() *Track {
	queueMutex.Lock()
	defer queueMutex.Unlock()
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

// SkipCurrent immediately skips the current track and pops the next.
func SkipCurrent() *Track {
	return PopNext()
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
