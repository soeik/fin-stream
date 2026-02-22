package delivery

import (
	"finstream/engine/internal/market"
	"finstream/engine/internal/market/mktproto"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (h *Hub) writeTrade(conn *websocket.Conn, stats []market.TickStats, format string) error {
	if format == "proto" {
		protoBatch := &mktproto.MarketSnapshot{
			Stats: make([]*mktproto.TickStats, len(stats)),
		}

		for i, s := range stats {
			protoBatch.Stats[i] = &mktproto.TickStats{
				Symbol:        s.Symbol,
				Price:         s.Price.String(),
				AvgPrice:      s.AvgPrice.String(),
				MinPrice:      s.MinPrice.String(),
				MaxPrice:      s.MaxPrice.String(),
				IsVolumeSpike: s.IsVolumeSpike,
			}
		}

		binaryData, err := proto.Marshal(protoBatch)
		if err != nil {
			return err
		}

		return conn.WriteMessage(websocket.BinaryMessage, binaryData)
	}
	// FIXME Doesn't seem to work after refactoring
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

	clientChan := make(chan []market.TickStats, 100)
	h.Register(clientChan)
	defer h.Unregister(clientChan)

	log.Println("New client connected")

	for stats := range clientChan {
		if err := h.writeTrade(conn, stats, format); err != nil {
			log.Printf("Delivery error [%s]: %v", format, err)
			break
		}
	}
}
