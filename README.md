# stream-chatbot

A Go-powered Twitch chatbot with a music request system and OBS browser overlay.

## Features

- **Twitch Chat Commands** — Polls, quotes, alphabetizer, and more
- **Music Request System** — Viewers request YouTube songs via chat; audio plays through an OBS browser source overlay
- **Real-time Overlay** — WebSocket-driven "Now Playing" overlay with animated EQ visualizer, track info, and transparent background for OBS
- **Rate Limiting** — Configurable cooldown to prevent spam (2 requests per 10 minutes per user)

## Music Request System

### How It Works

1. The chatbot listens for `!request` commands in Twitch chat
2. When a viewer requests a song, the bot searches the YouTube Data API and adds the top result to a queue
3. A local web server (port `38080`) serves an HTML overlay page at `http://localhost:38080/music`
4. The overlay connects via WebSocket and receives real-time state updates (current track, queue, pause state)
5. The YouTube IFrame API plays audio in the browser source — OBS captures the audio output

### Music Request Commands

All music commands use the `!request` prefix:

| Command | Who | Description |
|---------|-----|-------------|
| `!request add <query or URL>` | Everyone | Search YouTube and add the top result to the queue |
| `!request search <query>` | Everyone | Search YouTube and show up to 3 results (use `!request add <number>` to queue one) |
| `!request add <number>` | Everyone | Add a result from your last search by its number |
| `!request remove [query\|number]` | Everyone* | Remove a track from the queue by title, index, or most recent. *Non-mods can only remove their own requests |
| `!request queue` | Everyone | Show the next 3 tracks in the queue |
| `!request info` | Everyone | Show details about the currently playing track |
| `!request play` | Mods | Start playback or resume from pause |
| `!request pause` | Mods | Pause the current track |
| `!request resume` | Mods | Resume the current track |
| `!request skip` / `done` | Mods | Skip the current track |
| `!request limit` | Mods | Toggle rate limiting on/off |

### General Chat Commands

| Command | Who | Description |
|---------|-----|-------------|
| `!hello` / `!hellobot` | Everyone | Greets the user |
| `!bye` / `!byebot` | Everyone | Says goodbye to the user |
| `!quote` / `!randomquote` | Everyone | Displays a random Zen quote with a twist |
| `!abc <message>` / `!alpha <message>` | Everyone | Alphabetizes the words in your message |
| `!poll` / `!getPoll` | Everyone | Shows the results of the currently active Twitch poll |
| `!poll <question> // <choice 1> // <choice 2>` | Everyone | Creates a new Twitch poll with channel points voting (max 6 choices) |

### Music OBS Overlay Setup

1. Add a **Browser Source** in OBS
2. Set the URL to `http://localhost:38080/music`
3. Set the dimensions to **490 × 120**
4. The overlay has a transparent background — position it wherever you like on your scene
5. Make sure **"Control audio via OBS"** is configured to your preference (the YouTube player audio comes through the browser source)

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
│   ├── queue.go         # Track queue, add/remove/skip/pause logic, rate limiting
│   └── search.go        # YouTube Data API search and metadata fetching
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
