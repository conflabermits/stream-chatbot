let ws;
let currentTrackId = null;
let ytPlayer = null;
let lastPausedState = false;
let playTimeout = null;

// Initialize YouTube IFrame API
function onYouTubeIframeAPIReady() {
    ytPlayer = new YT.Player('yt-player', {
        height: '0', // Hide visually if we just want audio, or make small
        width: '0',
        playerVars: {
            'autoplay': 1,
            'controls': 0,
            'disablekb': 1,
            'fs': 0,
            'rel': 0
        },
        events: {
            'onReady': onPlayerReady,
            'onStateChange': onPlayerStateChange,
            'onError': onPlayerError
        }
    });
}

function onPlayerReady(event) {
    console.log("YouTube Player is ready");
    connectWebSocket();
}

function onPlayerStateChange(event) {
    if (event.data === YT.PlayerState.ENDED) {
        console.log("YouTube track ended");
        sendTrackEnded();
    }
}

function onPlayerError(event) {
    console.error("YouTube Player error:", event.data);
    // Auto-skip on error
    sendTrackEnded();
}

function connectWebSocket() {
    ws = new WebSocket("ws://localhost:38080/ws");

    ws.onopen = function() {
        console.log("Connected to WebSocket");
    };

    ws.onmessage = function(event) {
        const msg = JSON.parse(event.data);
        if (msg.event === "STATE_UPDATE") {
            handleStateUpdate(msg.data);
        }
    };

    ws.onclose = function() {
        console.log("WebSocket disconnected, retrying in 5s...");
        setTimeout(connectWebSocket, 5000);
    };
}

function handleStateUpdate(state) {
    const track = state.current_track;
    const upNext = state.up_next;
    const isPaused = state.is_paused;
    const container = document.getElementById('overlay-container');
    const musicIcon = document.querySelector('.music-icon');

    if (!track) {
        // Nothing actively playing — show "Up Next" or idle message
        stopAllActivePlayers();
        if (playTimeout) {
            clearTimeout(playTimeout);
        }
        currentTrackId = null;
        lastPausedState = false;

        if (upNext) {
            // Show "Up Next" screen
            document.getElementById('track-title').innerText = "Up Next: " + upNext.title;
            document.getElementById('track-artist').innerText = upNext.artist;
            document.getElementById('track-requester').innerText = "Requested by " + upNext.requested_by;
            musicIcon.classList.add('paused');
        } else {
            document.getElementById('track-title').innerText = "Nothing currently playing";
            document.getElementById('track-artist').innerText = "Type '!SL' or '!REQUEST' to add a song!";
            document.getElementById('track-requester').innerText = "";
            musicIcon.classList.add('paused');
        }
        return;
    }

    // A track is active
    if (currentTrackId !== track.id) {
        // New track started
        currentTrackId = track.id;
        lastPausedState = false;

        // Update UI
        document.getElementById('track-title').innerText = track.title;
        document.getElementById('track-artist').innerText = track.artist;
        document.getElementById('track-requester').innerText = "Requested by " + track.requested_by;
        container.classList.remove('hidden');

        stopAllActivePlayers();
        if (playTimeout) {
            clearTimeout(playTimeout);
        }
        
        playTimeout = setTimeout(() => {
            playTrack(track);
        }, 4000);
    }

    // Handle pause/resume state changes
    if (isPaused !== lastPausedState) {
        lastPausedState = isPaused;
        if (ytPlayer && typeof ytPlayer.pauseVideo === 'function') {
            if (isPaused) {
                ytPlayer.pauseVideo();
            } else {
                ytPlayer.playVideo();
            }
        }
    }

    // Update music icon animation based on pause state
    if (isPaused) {
        musicIcon.classList.add('paused');
    } else {
        musicIcon.classList.remove('paused');
    }
}

function stopAllActivePlayers() {
    if (ytPlayer && typeof ytPlayer.stopVideo === 'function') {
        ytPlayer.stopVideo();
    }
}

function playTrack(track) {
    stopAllActivePlayers();

    if (ytPlayer && typeof ytPlayer.loadVideoById === 'function') {
        ytPlayer.loadVideoById(track.id);
        ytPlayer.playVideo();
    }
}

function sendTrackEnded() {
    if (ws && ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({ event: "TRACK_ENDED" }));
    }
}

// Fallback in case YouTube API fails to load
setTimeout(() => {
    if (!ytPlayer) {
        console.warn("YouTube API timeout. Proceeding without it.");
        connectWebSocket();
    }
}, 3000);
