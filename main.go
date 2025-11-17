package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	port := flag.Int("port", 8088, "port to bind on 127.0.0.1")
	disableAutoStart := flag.Bool("no-autostart", false, "do not register HKCU Run on startup")
	openUI := flag.Bool("open-ui", false, "open web UI in the default browser on start")
	flag.Parse()

	addr := fmt.Sprintf("127.0.0.1:%d", *port)
	if !*disableAutoStart {
		if err := EnsureAutoStart(); err != nil {
			log.Printf("autostart: %v", err)
		}
	}

	handler := buildHandler()
	server := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	go func() {
		log.Printf("HTTP UI/API available at http://%s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server error: %v", err)
		}
	}()

	shutdown := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("http shutdown: %v", err)
		}
	}

	quitFromSignal := make(chan struct{})
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		<-sig
		close(quitFromSignal)
		shutdown()
		// Also quit the tray loop if it's running.
		systrayQuit()
	}()

	if *openUI {
		go openBrowser("http://" + addr)
	}

	runTray(trayConfig{
		address:      addr,
		openBrowser:  func() { openBrowser("http://" + addr) },
		shutdown:     shutdown,
		shutdownChan: quitFromSignal,
	})
}
