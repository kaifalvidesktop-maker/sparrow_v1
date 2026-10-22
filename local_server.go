package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// LocalHTMLServer serves a user-selected HTML document and its sibling assets
// from a localhost-only endpoint. This makes double-clicked .html files behave
// like normal browser pages while preserving relative CSS/JS/image paths.
type LocalHTMLServer struct {
	mu           sync.Mutex
	server       *http.Server
	rootDir      string
	filePath     string
	url          string
	ownedRootDir bool
}

var localHTMLServer = &LocalHTMLServer{}

func (s *LocalHTMLServer) StartFromContents(fileName string, content string) (string, error) {
	trimmedName := strings.TrimSpace(fileName)
	if trimmedName == "" {
		trimmedName = "index.html"
	}
	baseName := filepath.Base(trimmedName)
	if baseName == "." || baseName == string(filepath.Separator) || baseName == "" {
		baseName = "index.html"
	}
	ext := strings.ToLower(filepath.Ext(baseName))
	if ext != ".html" && ext != ".htm" {
		return "", fmt.Errorf("only html files are supported")
	}

	dir, err := os.MkdirTemp("", "sparrow-local-html-")
	if err != nil {
		return "", err
	}
	fullPath := filepath.Join(dir, baseName)
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		_ = os.RemoveAll(dir)
		return "", err
	}
	s.mu.Lock()
	// Mark the temporary directory as owned by this server before Start takes over.
	s.ownedRootDir = true
	s.mu.Unlock()
	return s.Start(fullPath)
}

func (s *LocalHTMLServer) Start(filePath string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.server != nil {
		s.stopLocked()
	}

	cleanPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(cleanPath)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return "", fmt.Errorf("local html must point to a file")
	}

	ext := strings.ToLower(filepath.Ext(cleanPath))
	if ext != ".html" && ext != ".htm" {
		return "", fmt.Errorf("local html must use .html or .htm")
	}

	rootDir := filepath.Dir(cleanPath)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", err
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		host := strings.ToLower(r.Host)
		if host != "localhost" && host != "127.0.0.1" && host != "[::1]" &&
			!strings.HasPrefix(host, "127.0.0.1:") &&
			!strings.HasPrefix(host, "localhost:") &&
			!strings.HasPrefix(host, "[::1]:") {
			http.Error(w, "Forbidden host", http.StatusForbidden)
			return
		}

		requestPath := r.URL.Path
		if requestPath == "" || requestPath == "/" {
			requestPath = "/" + filepath.Base(cleanPath)
		}

		relative := filepath.FromSlash(strings.TrimPrefix(requestPath, "/"))
		candidate := filepath.Clean(filepath.Join(rootDir, relative))

		// Never allow the localhost endpoint to escape the selected HTML folder.
		rel, err := filepath.Rel(rootDir, candidate)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			http.Error(w, "Forbidden path", http.StatusForbidden)
			return
		}

		if info, err := os.Stat(candidate); err != nil || info.IsDir() {
			http.NotFound(w, r)
			return
		}

		http.ServeFile(w, r, candidate)
	})

	s.server = &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		_ = s.server.Serve(listener)
	}()

	port := listener.Addr().(*net.TCPAddr).Port
	s.rootDir = rootDir
	s.filePath = cleanPath
	s.url = "http://127.0.0.1:" + strconv.Itoa(port) + "/"
	return s.url, nil
}

func (s *LocalHTMLServer) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stopLocked()
}

func (s *LocalHTMLServer) stopLocked() {
	if s.server == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = s.server.Shutdown(ctx)
	s.server = nil
	if s.ownedRootDir && s.rootDir != "" {
		_ = os.RemoveAll(s.rootDir)
	}
	s.rootDir = ""
	s.filePath = ""
	s.url = ""
	s.ownedRootDir = false
}

func (s *LocalHTMLServer) URL() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.url
}
