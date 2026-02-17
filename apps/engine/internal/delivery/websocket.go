package delivery

import (
	"finstream/engine/internal/models"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (h *Hub) writeTrade(conn *websocket.Conn, stats []models.TradeStats, format string) error {
	if format == "proto" {

		protoBatch := &models.MarketUpdateProto{
			Stats: make([]*models.TradeStatsProto, len(stats)),
		}

		for i, s := range stats {
			protoBatch.Stats[i] = &models.TradeStatsProto{
				Symbol:   s.Symbol,
				Price:    s.Price,
				AvgPrice: s.AvgPrice,
			}
		}

		binaryData, err := proto.Marshal(protoBatch)
		if err != nil {
			return err
		}

		return conn.WriteMessage(websocket.BinaryMessage, binaryData)
	}

	return conn.WriteJSON(stats)
}

func (h *Hub) HandleWS(w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WS upgrade error: %v", err)
		return
	}
	defer conn.Close()

	clientChan := make(chan []models.TradeStats, 100)
	h.Register(clientChan)
	defer h.Unregister(clientChan)

	log.Println("🌐 New client connected")

	for stats := range clientChan {
		if err := h.writeTrade(conn, stats, format); err != nil {
			log.Printf("Delivery error [%s]: %v", format, err)
			break
		}
	}
}
