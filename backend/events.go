package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

type EventHub struct {
	mu      sync.Mutex
	clients map[chan []byte]struct{}
	last    []byte
}

func NewEventHub() *EventHub { return &EventHub{clients: map[chan []byte]struct{}{}} }

func (h *EventHub) Publish(v any) {
	b, _ := json.Marshal(v)
	h.mu.Lock()
	defer h.mu.Unlock()
	h.last = b
	for ch := range h.clients {
		select {
		case ch <- b:
		default:
		}
	}
}

func (h *EventHub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming no soportado", 500)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	ch := make(chan []byte, 16)
	h.mu.Lock()
	h.clients[ch] = struct{}{}
	last := h.last
	h.mu.Unlock()
	defer func() { h.mu.Lock(); delete(h.clients, ch); h.mu.Unlock(); close(ch) }()
	if len(last) > 0 {
		fmt.Fprintf(w, "data: %s\n\n", last)
		f.Flush()
	}
	for {
		select {
		case b := <-ch:
			fmt.Fprintf(w, "data: %s\n\n", b)
			f.Flush()
		case <-r.Context().Done():
			return
		}
	}
}
