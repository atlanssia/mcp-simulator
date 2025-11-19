package transport

import (
	"fmt"
	"net/http"
)

// SSEAdapter handles Server-Sent Events connections.
type SSEAdapter struct{}

func NewSSEAdapter() *SSEAdapter {
	return &SSEAdapter{}
}

func (s *SSEAdapter) HandleSSE(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	// Send initial endpoint event as per MCP spec
	fmt.Fprintf(w, "event: endpoint\ndata: /messages\n\n")
	flusher.Flush()

	// Keep connection open
	<-r.Context().Done()
}
