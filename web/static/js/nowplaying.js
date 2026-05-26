let ws;
let currentTrackId = null;
let hideTimeout = null;

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
    const container = document.getElementById('fullscreen-container');

    if (!track) {
        // Nothing playing, hide immediately without restarting timeout
        if (container.classList.contains('show')) {
            hideContainer();
        }
        currentTrackId = null;
        return;
    }

    if (currentTrackId !== track.id) {
        currentTrackId = track.id;

        // Update DOM elements
        // YouTube API returns high-res thumbnails at maxresdefault.jpg or hqdefault.jpg
        // The bot sends the default url. We can try to upres it.
        let thumbUrl = track.thumbnail;
        if (thumbUrl && thumbUrl.includes('default.jpg')) {
            // Try to use a higher resolution thumbnail
            thumbUrl = thumbUrl.replace('default.jpg', 'maxresdefault.jpg');
        }

        document.getElementById('track-thumbnail').src = thumbUrl || '';
        document.getElementById('track-title').innerText = track.title;
        document.getElementById('track-artist').innerText = track.artist;
        document.getElementById('track-requester').innerText = "Requested by " + track.requested_by;

        // Reset any existing hide timeout
        if (hideTimeout) {
            clearTimeout(hideTimeout);
        }

        // Show the container
        container.classList.add('show');

        // Hide after 3 seconds
        hideTimeout = setTimeout(() => {
            hideContainer();
        }, 3000);
    }
}

function hideContainer() {
    const container = document.getElementById('fullscreen-container');
    container.classList.remove('show');
}

// Fallback error handler for high-res thumbnail (maxresdefault might not exist)
document.getElementById('track-thumbnail').addEventListener('error', function(e) {
    if (this.src.includes('maxresdefault.jpg')) {
        this.src = this.src.replace('maxresdefault.jpg', 'hqdefault.jpg');
    }
});

connectWebSocket();
