package channel

import (
	"sync"

	"github.com/apex/log"
)

type Handler struct {
	mu sync.Mutex
	ch chan<- string
}

func New(ch chan<- string) *Handler {
	return &Handler{ch: ch}
}

func (h *Handler) HandleLog(e *log.Entry) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	msg := e.Message

	if e.Level == log.ErrorLevel {
		msg = msg + " : ERROR : " + e.Fields.Get("error").(string)
	}

	h.ch <- msg
	return nil
}
