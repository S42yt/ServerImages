package handlers

import (
	"bufio"
	"encoding/json"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
)

// eventHub fans out server events to all connected SSE clients.
type eventHub struct {
	mu      sync.Mutex
	clients map[chan string]struct{}
}

var hub = &eventHub{clients: make(map[chan string]struct{})}

func (h *eventHub) add() chan string {
	ch := make(chan string, 8)
	h.mu.Lock()
	h.clients[ch] = struct{}{}
	h.mu.Unlock()
	return ch
}

func (h *eventHub) remove(ch chan string) {
	h.mu.Lock()
	if _, ok := h.clients[ch]; ok {
		delete(h.clients, ch)
		close(ch)
	}
	h.mu.Unlock()
}

// broadcast marshals data and pushes an SSE message to every client.
// ponytail: bounded per-client buffer; a slow client drops events rather than
// blocking everyone. Switch to per-client backpressure only if drops matter.
func (h *eventHub) broadcast(event string, data any) {
	payload, err := json.Marshal(data)
	if err != nil {
		return
	}
	msg := "event: " + event + "\ndata: " + string(payload) + "\n\n"

	h.mu.Lock()
	for ch := range h.clients {
		select {
		case ch <- msg:
		default:
		}
	}
	h.mu.Unlock()
}

// Events streams upload/delete events to clients over Server-Sent Events.
func Events() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/event-stream")
		c.Set("Cache-Control", "no-cache")
		c.Set("Connection", "keep-alive")

		ch := hub.add()

		c.Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
			defer hub.remove(ch)

			if _, err := w.WriteString(": connected\n\n"); err != nil {
				return
			}
			if err := w.Flush(); err != nil {
				return
			}

			// Keepalive pings double as dead-connection detection: a failed
			// write means the client is gone, so we exit and clean up.
			ticker := time.NewTicker(25 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case msg, ok := <-ch:
					if !ok {
						return
					}
					if _, err := w.WriteString(msg); err != nil {
						return
					}
					if err := w.Flush(); err != nil {
						return
					}
				case <-ticker.C:
					if _, err := w.WriteString(": ping\n\n"); err != nil {
						return
					}
					if err := w.Flush(); err != nil {
						return
					}
				}
			}
		}))

		return nil
	}
}
