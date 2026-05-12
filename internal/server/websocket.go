package server

import (
"encoding/json"
"log/slog"
"net/http"

"github.com/gin-gonic/gin"
"github.com/gorilla/websocket"
"github.com/Matthieusz/AVMS/internal/pqc"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for PoC
	},
}

type V2XMessage struct {
	CarID   string `json:"carId"`
	Type    string `json:"type"` // "AUTH_REQ", "AUTH_RES", "KEM_REQ", "KEM_RES"
	Payload string `json:"payload"`
}

func (s *Server) wsHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		slog.Error("websocket upgrade error", "error", err)
		return
	}
	defer conn.Close()

	slog.Info("New WebSocket connection established")

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			slog.Info("WebSocket disconnected", "error", err)
			break
		}

		var req V2XMessage
		if err := json.Unmarshal(message, &req); err != nil {
			slog.Error("invalid websocket message", "error", err)
			continue
		}

		slog.Info("Received message", "car", req.CarID, "type", req.Type)

		// Extended protocol for the PoC
		switch req.Type {
		case "AUTH_REQ":
			// Respond with Challenge
			res := V2XMessage{
				CarID:   req.CarID,
				Type:    "AUTH_CHALLENGE",
				Payload: "RSU_READY_PQC",
			}
			conn.WriteJSON(res)
		case "KEM_REQ":
			// Perform post-quantum KEM
			resText := "Shared secret established via PQC"
			resKEM, err := pqc.RunKEMCheck("Kyber512")
			if err == nil && resKEM.SharedSecretsCoincide {
				resText = "KEM Match: " + resKEM.KEMName
			}

			res := V2XMessage{
				CarID:   req.CarID,
				Type:    "KEM_SUCCESS",
				Payload: resText,
			}
			conn.WriteJSON(res)
		case "CAM":
			// Cooperative Awareness Message (Position, speed)
			// Server could process the telemetry here
		case "DENM":
			// Decentralized Environmental Notification (e.g., hard braking, hazards)
			// Server can broadcast hazards back to other vehicles
			res := V2XMessage{
				CarID:   "RSU-1",
				Type:    "HAZARD_WARNING",
				Payload: "Hazard near " + req.CarID,
			}
			conn.WriteJSON(res)
		}
	}
}
