package main

import (
	"fmt"
	"time"

	"github.com/marcusolsson/tui-go"
)

func main() {
	downloads := tui.NewList()

	dlBox := tui.NewVBox(downloads)
	dlBox.SetTitle("downloads")

	logs := tui.NewVBox()

	logScroll := tui.NewScrollArea(logs)
	logBox := tui.NewVBox(logScroll)
	logBox.SetTitle("logs")

	root := tui.NewHBox(logBox, dlBox)

	ui, err := tui.New(root)
	if err != nil {
		panic(err)
	}
	ui.SetKeybinding("Esc", func() { ui.Quit() })

	go func() {
		for {
			logs.Append(tui.NewHBox(
				tui.NewLabel(time.Now().Format(time.RFC3339)),
				tui.NewSpacer(),
			))
			ui.Update(func() {})
			time.Sleep(time.Second * 5)
		}
	}()

	go func() {
		movies := []string{
			"%s - movie #1",
			"%s - movie #2",
		}
		for {
			downloads.RemoveItems()
			for _, movie := range movies {
				downloads.AddItems(fmt.Sprintf(movie, time.Now().Format("05.000")))
			}
			ui.Update(func() {})
			time.Sleep(time.Second * 15)
		}
	}()

	if err := ui.Run(); err != nil {
		panic(err)
	}
}
