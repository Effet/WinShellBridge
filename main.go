package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	configPath := flag.String("config", defaultConfigPath(), "path to JSON config file")
	port := flag.Int("port", 8088, "port to bind on 127.0.0.1")
	host := flag.String("host", "127.0.0.1", "host/IP to bind (set to 0.0.0.0 to listen on all interfaces)")
	disableAutoStart := flag.Bool("no-autostart", false, "do not register HKCU Run on startup")
	openUI := flag.Bool("open-ui", false, "open web UI in the default browser on start")
	flag.Parse()

	visited := map[string]bool{}
	flag.CommandLine.Visit(func(f *flag.Flag) {
		visited[f.Name] = true
	})

	if cfg, err := loadConfig(*configPath); err != nil {
		log.Printf("config load failed: %v", err)
	} else if cfg != nil {
		if cfg.Host != "" && !visited["host"] {
			*host = cfg.Host
		}
		if cfg.Port != 0 && !visited["port"] {
			*port = cfg.Port
		}
		if cfg.Autostart != nil && !visited["no-autostart"] {
			*disableAutoStart = !*cfg.Autostart
		}
		if cfg.OpenUI != nil && !visited["open-ui"] {
			*openUI = *cfg.OpenUI
		}
		log.Printf("config loaded from %s", *configPath)
	}

	addr := net.JoinHostPort(*host, fmt.Sprintf("%d", *port))
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

	openHost := *host
	if openHost == "" || openHost == "0.0.0.0" || openHost == "::" {
		openHost = "127.0.0.1"
	}
	openURL := "http://" + net.JoinHostPort(openHost, fmt.Sprintf("%d", *port))

	if *openUI {
		go openBrowser(openURL)
	}

	runTray(trayConfig{
		address:      addr,
		openBrowser:  func() { openBrowser(openURL) },
		shutdown:     shutdown,
		shutdownChan: quitFromSignal,
	})
}
