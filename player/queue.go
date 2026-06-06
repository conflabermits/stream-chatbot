package player

import (
	"errors"
	"sync"
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
	currentTrack *Track
	queueMutex   sync.Mutex
	IsPaused     bool

	AutoplayEnabled = false

	OnTrackEnded func()
)

// NotifyTrackEnded is called when a track finishes playing naturally or is skipped.
func NotifyTrackEnded() {
	if OnTrackEnded != nil {
		OnTrackEnded()
	}
}

// ToggleAutoplay toggles the autoplay feature.
func ToggleAutoplay() bool {
	queueMutex.Lock()
	defer queueMutex.Unlock()
	AutoplayEnabled = !AutoplayEnabled
	return AutoplayEnabled
}

// GetAutoplay returns whether autoplay is enabled.
func GetAutoplay() bool {
	queueMutex.Lock()
	defer queueMutex.Unlock()
	return AutoplayEnabled
}

// ClearCurrent stops the currently playing track and leaves the player waiting.
func ClearCurrent() {
	queueMutex.Lock()
	defer queueMutex.Unlock()
	currentTrack = nil
	IsPaused = false
}

// PlayOrResume resumes playback if paused, or fetches next track from SSL if stopped.
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

	// Fetch next from SSL
	track, err := FetchNextFromSSL()
	if err != nil {
		return err
	}

	currentTrack = track
	IsPaused = false
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

// GetCurrentTrack returns the currently playing track.
func GetCurrentTrack() *Track {
	queueMutex.Lock()
	defer queueMutex.Unlock()
	return currentTrack
}

// GetQueue is retained for compatibility with web overlay websocket.
func GetQueue() []Track {
	return []Track{}
}
