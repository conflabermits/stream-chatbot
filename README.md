# stream-chatbot
A repository specifically for the Go-powered Twitch chatbot I intend to use during streams.

## Features

### Twitch Chatbot
The bot connects to your Twitch channel and listens for commands. Built-in commands include:
* `!hello` / `!hellobot` - A friendly greeting.
* `!quote` - A random inspirational quote.
* `!poll` - Create polls directly from the chat.

### Spoopify (Spotify Integration)
The bot includes a robust Spotify integration allowing viewers to query tracks and letting moderators control the stream's music queue.

> **Note:** Accessing the Spotify Player Web API (used by `queue`, `add`, `pause`, `play`, `next`) requires the authenticated user to have an active **Spotify Premium** subscription.

#### Spotify Commands:
* `!spoopify search <query>`: Search for songs and return the top 5 results with artist, title, length, and unique ID.
* `!spoopify info <song ID>`: Return detailed info on a specific song, including album and release year.
* `!spoopify add <song ID or query>`: Add a song to the currently playing queue.
* `!spoopify queue` or `!spoopify list`: Show the next 10 songs in the current playback queue.
* `!spoopify pause`: Pause playback (Broadcaster/Mod only).
* `!spoopify play` / `!spoopify resume`: Resume playback (Broadcaster/Mod only).
* `!spoopify next` / `!spoopify skip`: Skip to the next track (Broadcaster/Mod only).

## Setup & Configuration

The application expects a `.creds` file in the root directory (see `example.creds` for a template). 

### Spotify Authentication
To use the `!spoopify` commands:
1. Register an application in the [Spotify Developer Dashboard](https://developer.spotify.com/dashboard/).
2. Set the Redirect URI in the dashboard to `http://127.0.0.1:8081/callback`. *(Note: Spotify strictly requires HTTPS for most URIs, but explicitly allows `http://127.0.0.1` for local development).*
3. Add `SpotifyClientID` and `SpotifyClientSecret` to your `.creds` file.
4. Run the application. If a valid token isn't found, it will host a local server on `http://127.0.0.1:8081/`. Navigate there in your browser to complete the Spotify OAuth flow.

## Local Testing
You can interact with the bot directly via standard input (console) while it is running. The bot will simulate these inputs as if they were sent by the broadcaster in Twitch chat, allowing you to test commands without needing to be live on Twitch.
