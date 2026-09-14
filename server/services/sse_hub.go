package services

import (
	"encoding/json"
	"sync"
)

type PharmacyAlertEvent struct {
	Type  string      `json:"type"`
	Alert interface{} `json:"alert,omitempty"`
	AlertID int32     `json:"alert_id,omitempty"`
}

type CustomerAlertEvent struct {
	Type   string      `json:"type"`
	AlertID int32      `json:"alert_id,omitempty"`
	Data   interface{} `json:"data,omitempty"`
}

type SSEHub struct {
	mu sync.RWMutex

	pharmacyClients map[int32]map[chan string]struct{}
	alertClients    map[int32]map[chan string]struct{}
}

func NewSSEHub() *SSEHub {
	return &SSEHub{
		pharmacyClients: make(map[int32]map[chan string]struct{}),
		alertClients:    make(map[int32]map[chan string]struct{}),
	}
}

func (h *SSEHub) RegisterPharmacy(pharmacyID int32, ch chan string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.pharmacyClients[pharmacyID] == nil {
		h.pharmacyClients[pharmacyID] = make(map[chan string]struct{})
	}
	h.pharmacyClients[pharmacyID][ch] = struct{}{}
}

func (h *SSEHub) UnregisterPharmacy(pharmacyID int32, ch chan string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, ok := h.pharmacyClients[pharmacyID]; ok {
		delete(clients, ch)
		if len(clients) == 0 {
			delete(h.pharmacyClients, pharmacyID)
		}
	}
	close(ch)
}

func (h *SSEHub) BroadcastToPharmacy(pharmacyID int32, event PharmacyAlertEvent) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients, ok := h.pharmacyClients[pharmacyID]
	if !ok || len(clients) == 0 {
		return
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return
	}
	msg := string(payload)

	for ch := range clients {
		select {
		case ch <- msg:
		default:
		}
	}
}

func (h *SSEHub) RegisterAlert(alertID int32, ch chan string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.alertClients[alertID] == nil {
		h.alertClients[alertID] = make(map[chan string]struct{})
	}
	h.alertClients[alertID][ch] = struct{}{}
}

func (h *SSEHub) UnregisterAlert(alertID int32, ch chan string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, ok := h.alertClients[alertID]; ok {
		delete(clients, ch)
		if len(clients) == 0 {
			delete(h.alertClients, alertID)
		}
	}
	close(ch)
}

func (h *SSEHub) BroadcastToAlert(alertID int32, event CustomerAlertEvent) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients, ok := h.alertClients[alertID]
	if !ok || len(clients) == 0 {
		return
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return
	}
	msg := string(payload)

	for ch := range clients {
		select {
		case ch <- msg:
		default:
		}
	}
}
