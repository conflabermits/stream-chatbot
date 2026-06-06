# stream-chatbot

A Go-powered Twitch chatbot with a music request system and OBS browser overlay.

## Features

- **Twitch Chat Commands** — Polls, quotes, alphabetizer, and more
- **Streamer Songlist Integration** — Automatically fetches the next song from your Streamer Songlist queue and plays it through an OBS browser source overlay
- **Real-time Overlay** — WebSocket-driven "Now Playing" overlay with animated EQ visualizer, track info, and transparent background for OBS

## Music Request System

### How It Works

1. The bot interfaces with the Streamer Songlist API to fetch the top song from your channel's queue
2. A local web server (port `38080`) serves an HTML overlay page at `http://localhost:38080/music`
3. The overlay connects via WebSocket and receives real-time state updates (current track, pause state)
4. The YouTube IFrame API plays audio in the browser source — OBS captures the audio output
5. When a track finishes, the bot sends `!setPlayed` in Twitch chat to automatically advance your Streamer Songlist queue

### Music Request Commands

All music commands use the `!request` prefix:

| Command | Alias | Who | Description |
|---------|-------|-----|-------------|
| `!request info` | `!info` | Everyone | Show details about the currently playing track |
| `!request play` | `!play` | Mods | Start playback or resume from pause (fetches next track from Streamer Songlist) |
| `!request pause` | `!pause` | Mods | Pause the current track |
| `!request resume` | `!resume` | Mods | Resume the current track |
| `!request skip` / `done` | `!skip` / `!done` | Mods | Skip the current track (automatically plays the next track if autoplay is enabled) |
| `!request autoplay` | `!autoplay` | Mods | Toggle autoplaying the next song in the queue |

### General Chat Commands

| Command | Who | Description |
|---------|-----|-------------|
| `!hello` / `!hellobot` | Everyone | Greets the user |
| `!bye` / `!byebot` | Everyone | Says goodbye to the user |
| `!quote` / `!randomquote` | Everyone | Displays a random Zen quote with a twist |
| `!abc <message>` / `!alpha <message>` | Everyone | Alphabetizes the words in your message |
| `!poll` / `!getPoll` | Everyone | Shows the results of the currently active Twitch poll |
| `!poll <question> // <choice 1> // <choice 2>` | Everyone | Creates a new Twitch poll with channel points voting (max 6 choices) |

### Music OBS Overlays Setup

The bot provides two distinct browser overlays for music:

**1. Mini Player & Audio Source (Required)**
This is the main player that handles the audio and displays a small "Now Playing" widget.
- Add a **Browser Source** in OBS
- Set the URL to `http://localhost:38080/music`
- Set the dimensions to **490 × 120**
- The overlay has a transparent background — position it wherever you like on your scene
- Make sure **"Control audio via OBS"** is configured to your preference (the YouTube player audio comes through the browser source)

**2. Full-Screen "Now Playing" Splash (Optional)**
This overlay stays 100% transparent but displays a large, centered splash screen with the album art and track info for 3 seconds whenever a new song starts playing. Note: The audio in the Mini Player is intentionally delayed by 4 seconds to allow this splash screen animation to finish gracefully.
- Add a **Browser Source** in OBS
- Set the URL to `http://localhost:38080/nowplaying`
- Set the dimensions to match your stream output (e.g., **1920 × 1080**)
- Place it above your game capture, but behind critical alerts
- **Opacity Tweak**: To adjust how transparent this overlay is, add `opacity: 0.9;` (or your preferred value) to the OBS browser source's Custom CSS field within the `body { ... }` block.

## Donorbox Overlay Setup

The bot also hosts a web overlay for Donorbox progress, checking a specified campaign page automatically.

1. Add a **Browser Source** in OBS
2. Set the URL to `http://localhost:28080/donorbox`
3. Set the dimensions as needed for your stream layout
4. Run the chatbot with the `--url` flag to monitor a specific campaign (e.g. `./stream-chatbot --url https://donorbox.org/your-campaign`). It checks the page every 60 seconds by default.

## Configuration

Copy `example.creds` to `.creds` and fill in your values:

```
ClientID=<your twitch app client id>
ClientSecret=<your twitch app client secret>
TwitchUsername=<bot account username>
TwitchChannel=<your channel name>
BroadcasterID=<your broadcaster id>
TwitchToken=<twitch oauth token>
YouTubeAPIKey=<youtube data api v3 key>
```

The `YouTubeAPIKey` is required for the music request system. You can get one from the [Google Cloud Console](https://console.cloud.google.com/) by enabling the YouTube Data API v3.

## Project Structure

```
stream-chatbot/
├── main.go              # Entry point — starts chatbot, auth server, and overlay server
├── auth/                # Twitch OAuth token management
├── chatbot/chatbot.go   # Twitch IRC bot with all chat command handlers
├── common/common.go     # Shared credentials loader
├── player/
│   ├── queue.go         # Track state, skip/pause logic, autoplay setting
│   ├── search.go        # YouTube Data API search and metadata fetching
│   └── ssl.go           # Streamer Songlist API integration to fetch tracks
├── web/
│   ├── music_overlay.go # HTTP server for the overlay page (port 38080)
│   ├── websocket.go     # WebSocket handler for real-time state broadcast
│   └── static/
│       ├── music.html   # Overlay HTML
│       ├── css/music.css # Overlay styles (transparent, OBS-ready)
│       └── js/music.js  # YouTube IFrame API player + WebSocket client
└── logs/                # Runtime log output
```

## Building & Running

```bash
go build -o stream-chatbot .
./stream-chatbot
```
