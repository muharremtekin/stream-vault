package handler

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"

	"github.com/streamvault/notification-service/internal/auth"
	"github.com/streamvault/notification-service/internal/config"
	ws "github.com/streamvault/notification-service/internal/websocket"
)

// WebSocketHandler handles WebSocket upgrade requests for notifications.
type WebSocketHandler struct {
	hub      *ws.Hub
	upgrader websocket.Upgrader
	jwtCfg   config.JWTConfig
	wsCfg    config.WebSocketConfig
}

// NewWebSocketHandler creates a new WebSocketHandler.
func NewWebSocketHandler(hub *ws.Hub, jwtCfg config.JWTConfig, wsCfg config.WebSocketConfig) *WebSocketHandler {
	return &WebSocketHandler{
		hub: hub,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  wsCfg.ReadBufferSize,
			WriteBufferSize: wsCfg.WriteBufferSize,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		jwtCfg: jwtCfg,
		wsCfg:  wsCfg,
	}
}

// ServeWS handles the GET /ws/notifications WebSocket upgrade.
func (h *WebSocketHandler) ServeWS(w http.ResponseWriter, r *http.Request) {
	auth, err := auth.ValidateWSToken(r, h.jwtCfg.Secret, h.jwtCfg.Issuer)
	if err != nil {
		log.Warn().Err(err).Msg("websocket auth failed")
		WriteErrorResponse(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "invalid or missing token")
		return
	}

	if h.hub.UserClientCount(auth.UserID) >= h.wsCfg.MaxConnectionsPerUser {
		log.Warn().
			Str("user_id", auth.UserID).
			Int("max", h.wsCfg.MaxConnectionsPerUser).
			Msg("max websocket connections per user reached")
		WriteErrorResponse(w, r, http.StatusTooManyRequests, "TOO_MANY_CONNECTIONS", "maximum websocket connections reached")
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("websocket upgrade failed")
		return
	}

	client := ws.NewClient(h.hub, conn, auth.UserID)
	h.hub.Register() <- client

	go client.WritePump(h.wsCfg.PingInterval)
	go client.ReadPump(h.wsCfg.PingInterval, h.wsCfg.PongTimeout)

	log.Info().
		Str("user_id", auth.UserID).
		Str("role", auth.Role).
		Msg("websocket client connected")
}
