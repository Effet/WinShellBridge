package main

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

//go:embed web/*
var webContent embed.FS

type runRequest struct {
	Cmd        string   `json:"cmd"`
	Args       []string `json:"args,omitempty"`
	Workdir    string   `json:"workdir,omitempty"`
	TimeoutSec int      `json:"timeout_sec,omitempty"`
	Background bool     `json:"background,omitempty"`
}

func buildHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/run", handleRun)

	// Web UI (embedded)
	root, err := fs.Sub(webContent, "web")
	if err != nil {
		log.Fatalf("embed error: %v", err)
	}
	fileServer := http.FileServer(http.FS(root))
	mux.Handle("/", fileServer)

	return logRequests(mux)
}

func handleRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()

	var req runRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid json: %v", err), http.StatusBadRequest)
		return
	}
	req.Cmd = strings.TrimSpace(req.Cmd)
	if req.Cmd == "" {
		http.Error(w, "cmd is required", http.StatusBadRequest)
		return
	}

	var ctx context.Context
	var cancel context.CancelFunc = func() {}
	if req.Background {
		ctx = context.Background()
		if req.TimeoutSec > 0 {
			ctx, cancel = context.WithTimeout(context.Background(), time.Duration(req.TimeoutSec)*time.Second)
		}
	} else {
		ctx = r.Context()
		if req.TimeoutSec > 0 {
			ctx, cancel = context.WithTimeout(ctx, time.Duration(req.TimeoutSec)*time.Second)
		}
	}
	defer cancel()

	cmd := exec.CommandContext(ctx, req.Cmd, req.Args...)
	if req.Workdir != "" {
		cmd.Dir = req.Workdir
	}
	// Best effort: normalize working dir so logs show clean path.
	if cmd.Dir != "" {
		if abs, err := filepath.Abs(cmd.Dir); err == nil {
			cmd.Dir = abs
		}
	}
	if req.Background {
		cmd.Stdout = io.Discard
		cmd.Stderr = io.Discard
	} else {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fw := &flushWriter{w: w}
		cmd.Stdout = fw
		cmd.Stderr = fw
	}

	if err := cmd.Start(); err != nil {
		http.Error(w, fmt.Sprintf("start failed: %v", err), http.StatusBadRequest)
		return
	}
	if req.Background {
		pid := cmd.Process.Pid
		go func() {
			if err := cmd.Wait(); err != nil && !errors.Is(err, context.Canceled) {
				log.Printf("background pid %d exited: %v", pid, err)
			}
		}()
		resp := map[string]any{
			"pid":               pid,
			"started":           true,
			"timeout_sec":       req.TimeoutSec,
			"background":        true,
			"working_directory": cmd.Dir,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	if err := cmd.Wait(); err != nil {
		fmt.Fprintf(w, "\n[exit error: %v]\n", err)
		return
	}
}

type flushWriter struct {
	w http.ResponseWriter
}

func (fw *flushWriter) Write(p []byte) (int, error) {
	n, err := fw.w.Write(p)
	if flusher, ok := fw.w.(http.Flusher); ok {
		flusher.Flush()
	}
	return n, err
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s (%s)", r.Method, r.URL.Path, time.Since(start).Truncate(time.Millisecond))
	})
}
