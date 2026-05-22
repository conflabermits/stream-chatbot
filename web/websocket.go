package overlay

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"stream-chatbot/player"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all for OBS browser source
	},
}

type WsMessage struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data,omitempty"`
}

type TrackState struct {
	CurrentTrack *player.Track  `json:"current_track"`
	Queue        []player.Track `json:"queue"`
}

var clients = make(map[*websocket.Conn]bool)
var clientsMutex sync.Mutex

func handleMusicWebSocket(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket Upgrade Error: %v", err)
		return
	}
	defer ws.Close()

	clientsMutex.Lock()
	clients[ws] = true
	clientsMutex.Unlock()

	// Send initial state
	sendStateToClient(ws)

	for {
		var msg WsMessage
		err := ws.ReadJSON(&msg)
		if err != nil {
			clientsMutex.Lock()
			delete(clients, ws)
			clientsMutex.Unlock()
			break
		}

		if msg.Event == "TRACK_ENDED" {
			log.Println("Received TRACK_ENDED from overlay, popping next track...")
			player.PopNext()
			BroadcastState()
		}
	}
}

func sendStateToClient(ws *websocket.Conn) {
	state := TrackState{
		CurrentTrack: player.GetCurrentTrack(),
		Queue:        player.GetQueue(),
	}
	msg := WsMessage{
		Event: "STATE_UPDATE",
		Data:  state,
	}
	_ = ws.WriteJSON(msg)
}

// BroadcastState sends the current track and queue to all connected overlay clients.
func BroadcastState() {
	state := TrackState{
		CurrentTrack: player.GetCurrentTrack(),
		Queue:        player.GetQueue(),
	}
	msg := WsMessage{
		Event: "STATE_UPDATE",
		Data:  state,
	}

	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	for client := range clients {
		err := client.WriteJSON(msg)
		if err != nil {
			client.Close()
			delete(clients, client)
		}
	}
}
