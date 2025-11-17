package main

import (
	"fmt"
	"log"

	"github.com/getlantern/systray"
)

type trayConfig struct {
	address      string
	openBrowser  func()
	shutdown     func()
	shutdownChan <-chan struct{}
}

func runTray(cfg trayConfig) {
	systray.Run(func() { onReady(cfg) }, func() {
		if cfg.shutdown != nil {
			cfg.shutdown()
		}
	})
}

func onReady(cfg trayConfig) {
	if len(trayIcon) > 0 {
		systray.SetIcon(trayIcon)
	}
	systray.SetTooltip("WinShellBridge - local shell bridge")
	systray.SetTitle("WinShellBridge")

	open := systray.AddMenuItem("Open UI", "Open the web UI")
	status := systray.AddMenuItem(fmt.Sprintf("Listening on %s", cfg.address), cfg.address)
	status.Disable()
	systray.AddSeparator()
	exit := systray.AddMenuItem("Quit", "Stop the bridge")

	go func() {
		for {
			select {
			case <-open.ClickedCh:
				if cfg.openBrowser != nil {
					cfg.openBrowser()
				}
			case <-exit.ClickedCh:
				systray.Quit()
				return
			case <-cfg.shutdownChan:
				log.Println("Shutting down on signal")
				systray.Quit()
				return
			}
		}
	}()
}

func systrayQuit() {
	systray.Quit()
}
