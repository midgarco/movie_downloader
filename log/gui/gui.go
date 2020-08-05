package gui

import (
	"github.com/apex/log"
	"github.com/jroimartin/gocui"
)

type Handler struct {
	lh log.Handler
	g  *gocui.Gui
}

func New(g *gocui.Gui, h log.Handler) *Handler {
	return &Handler{g: g, lh: h}
}

func (h *Handler) HandleLog(e *log.Entry) error {
	err := h.lh.HandleLog(e)
	if err != nil {
		return err
	}
	h.g.Update(func(*gocui.Gui) error { return nil })
	return nil
}
