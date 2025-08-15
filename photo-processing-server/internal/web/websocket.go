package web

import (
    "log"
    "net/http"
    "sync"

    "github.com/gin-gonic/gin"
    "github.com/gorilla/websocket"
    "photo-processing-server/internal/services"
    "photo-processing-server/internal/config"
)

// WebSocket message types matching TypeScript interfaces
type WSMessage struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

type ClientMessage struct {
	Type  string `json:"type"`
	JobID string `json:"jobId,omitempty"`
}

// WebSocket connection wrapper
type WSConnection struct {
	conn         *websocket.Conn
	send         chan WSMessage
	subscribedTo string // Job ID this connection is subscribed to
	userID       string // owner of the connection, may be empty if not logged in
}

// WebSocket hub manages all connections
type WSHub struct {
	connections map[*WSConnection]bool
	register    chan *WSConnection
	unregister  chan *WSConnection
	broadcast   chan WSMessage
	logger      *services.Logger
	mutex       sync.RWMutex
}

// Global hub instance
var hub *WSHub

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from localhost (in production, be more restrictive)
		return true
	},
}

// InitializeWebSocket sets up the WebSocket hub
func InitializeWebSocket(logger *services.Logger) {
	hub = &WSHub{
		connections: make(map[*WSConnection]bool),
		register:    make(chan *WSConnection),
		unregister:  make(chan *WSConnection),
		broadcast:   make(chan WSMessage),
		logger:      logger,
	}
	
	go hub.run()
	
	// Hook into logger to broadcast logs
	logger.SetWebSocketBroadcaster(func(message string) {
		// Do not broadcast raw logs globally; send only to clients subscribed to a job
		// Generic system logs without job context are ignored to prevent cross-user leak
	})
}

// Run the WebSocket hub
func (h *WSHub) run() {
	for {
		select {
		case conn := <-h.register:
			h.mutex.Lock()
			h.connections[conn] = true
			h.mutex.Unlock()
			log.Printf("WebSocket client connected, total: %d", len(h.connections))

		case conn := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.connections[conn]; ok {
				delete(h.connections, conn)
				close(conn.send)
			}
			h.mutex.Unlock()
			log.Printf("WebSocket client disconnected, total: %d", len(h.connections))

		case message := <-h.broadcast:
			h.mutex.RLock()
			for conn := range h.connections {
				// Only deliver messages with jobId to matching subscribers
				jobId, _ := message.Data["jobId"].(string)
				if jobId != "" && conn.subscribedTo == jobId {
					select {
					case conn.send <- message:
					default:
						delete(h.connections, conn)
						close(conn.send)
					}
				}
			}
			h.mutex.RUnlock()
		}
	}
}

// WebSocket handler
func HandleWebSocket(c *gin.Context) {
    cfg := config.Load()
    if cfg.APIToken != "" {
        auth := c.GetHeader("Authorization")
        if len(auth) < 8 || auth[:7] != "Bearer " || auth[7:] != cfg.APIToken {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
            return
        }
    }
    // Allow connect even without session (subscribe will enforce ownership)
    var uid string
    if cookie, err := c.Cookie("session_token"); err == nil {
        if u, ok := getUserIDBySession(cookie); ok {
            uid = u
        }
    }
    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        log.Printf("WebSocket upgrade error: %v", err)
        return
    }

    wsConn := &WSConnection{
        conn: conn,
        send: make(chan WSMessage, 256),
        userID: uid,
    }

    hub.register <- wsConn

    // Start goroutines for reading and writing
    go wsConn.writePump()
    go wsConn.readPump()
}

// Read from WebSocket
func (c *WSConnection) readPump() {
	defer func() {
		hub.unregister <- c
		c.conn.Close()
	}()

	for {
		var msg ClientMessage
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle client messages
		switch msg.Type {
		case "subscribe":
			// Validate ownership of the job
			if job, ok := GetJob(msg.JobID); ok {
				if c.userID != "" && job.UserID == c.userID {
					if c.subscribedTo != msg.JobID {
						c.subscribedTo = msg.JobID
						log.Printf("Client subscribed to job: %s", msg.JobID)
					}
					// Immediately send current job snapshot
					if job.Status == "completed" {
						c.send <- WSMessage{Type: "complete", Data: map[string]interface{}{"jobId": job.ID}}
					} else if job.Status == "error" {
						c.send <- WSMessage{Type: "error", Data: map[string]interface{}{"jobId": job.ID, "error": job.Error}}
					} else {
						c.send <- WSMessage{Type: "progress", Data: map[string]interface{}{"jobId": job.ID, "progress": job.Progress}}
					}
				}
			}
		case "unsubscribe":
			c.subscribedTo = ""
			log.Printf("Client unsubscribed")
		}
	}
}

// Write to WebSocket
func (c *WSConnection) writePump() {
	defer c.conn.Close()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.conn.WriteJSON(message)
		}
	}
}

// Broadcast functions for different message types

// BroadcastLog sends a log message to all connected clients
func BroadcastLog(message string) {
	if hub == nil {
		return
	}
	// Log messages are not broadcasted globally to prevent leaking between users
}

// BroadcastProgress sends progress update to subscribed clients
func BroadcastProgress(jobID string, progress float64) {
	if hub == nil {
		return
	}
	msg := WSMessage{
		Type: "progress",
		Data: map[string]interface{}{
			"progress": progress,
			"jobId":    jobID,
		},
	}
	hub.mutex.RLock()
	for conn := range hub.connections {
		if conn.subscribedTo == jobID {
			select {
			case conn.send <- msg:
			default:
				// Connection is blocked, skip
			}
		}
	}
	hub.mutex.RUnlock()
}

// BroadcastComplete sends completion message to subscribed clients
func BroadcastComplete(jobID string, result interface{}) {
	if hub == nil {
		return
	}
	msg := WSMessage{
		Type: "complete",
		Data: map[string]interface{}{
			"jobId":  jobID,
			"result": result,
		},
	}
	hub.mutex.RLock()
	for conn := range hub.connections {
		if conn.subscribedTo == jobID {
			select {
			case conn.send <- msg:
			default:
			}
		}
	}
	hub.mutex.RUnlock()
}

// BroadcastError sends error message to subscribed clients
func BroadcastError(jobID string, errorMsg string) {
	if hub == nil {
		return
	}
	msg := WSMessage{
		Type: "error",
		Data: map[string]interface{}{
			"jobId": jobID,
			"error": errorMsg,
		},
	}
	hub.mutex.RLock()
	for conn := range hub.connections {
		if conn.subscribedTo == jobID {
			select {
			case conn.send <- msg:
			default:
			}
		}
	}
	hub.mutex.RUnlock()
}

// Add WebSocket route to the router
func SetupWebSocketRoutes(router *gin.Engine) {
	router.GET("/ws", HandleWebSocket)
}