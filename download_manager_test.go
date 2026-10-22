package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestDownloadQueueLimitsConcurrentTransfers(t *testing.T) {
	downloadManager.ClearAll()
	t.Cleanup(ShutdownDownloads)

	var current atomic.Int32
	var maximum atomic.Int32
	release := make(chan struct{})
	var releaseOnce sync.Once
	releaseHandlers := func() { releaseOnce.Do(func() { close(release) }) }
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		active := current.Add(1)
		defer current.Add(-1)
		for observed := maximum.Load(); active > observed; observed = maximum.Load() {
			if maximum.CompareAndSwap(observed, active) {
				break
			}
		}
		<-release
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = io.WriteString(w, "download")
	}))
	defer server.Close()
	defer releaseHandlers()

	dir := t.TempDir()
	for i := 0; i < maxConcurrentDownloads+1; i++ {
		StartDownload(fmt.Sprintf("%s/file-%d.bin", server.URL, i), dir)
	}

	var queuedID int
	waitForDownloadState(t, func() bool {
		queued, running := 0, 0
		for _, item := range downloadManager.GetAll() {
			switch item.Status {
			case "queued":
				queued++
				queuedID = item.ID
			case "downloading":
				running++
			}
		}
		return queued == 1 && running == maxConcurrentDownloads
	})
	if !ToggleDownloadPause(queuedID) {
		t.Fatal("could not pause a queued download")
	}
	waitForDownloadState(t, func() bool {
		for _, item := range downloadManager.GetAll() {
			if item.ID == queuedID {
				return item.Status == "paused"
			}
		}
		return false
	})
	if ToggleDownloadPause(queuedID) {
		t.Fatal("could not resume a queued download")
	}
	waitForDownloadState(t, func() bool {
		return current.Load() == maxConcurrentDownloads
	})

	releaseHandlers()
	activeWG.Wait()
	if got := maximum.Load(); got != maxConcurrentDownloads {
		t.Fatalf("maximum concurrent downloads = %d, want %d", got, maxConcurrentDownloads)
	}
	for _, item := range downloadManager.GetAll() {
		if item.Status != "completed" {
			t.Errorf("download %d status = %q, want completed", item.ID, item.Status)
		}
	}
}

func TestDoDownloadResumesWithRangeRequest(t *testing.T) {
	const content = "Sparrow resumable download"
	var receivedRange atomic.Value
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rangeHeader := r.Header.Get("Range")
		receivedRange.Store(rangeHeader)
		if rangeHeader != "bytes=7-" {
			http.Error(w, "expected a range request", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Range", fmt.Sprintf("bytes 7-%d/%d", len(content)-1, len(content)))
		w.Header().Set("Content-Length", fmt.Sprint(len(content)-7))
		w.WriteHeader(http.StatusPartialContent)
		_, _ = io.WriteString(w, content[7:])
	}))
	defer server.Close()

	manager := &DownloadManager{NextID: 1}
	item := manager.Start(server.URL + "/package.bin")
	previousManager := downloadManager
	downloadManager = manager
	defer func() { downloadManager = previousManager }()

	dir := t.TempDir()
	finalPath := filepath.Join(dir, "package.bin")
	partialPath := finalPath + ".sparrow-part"
	if err := os.WriteFile(partialPath, []byte(content[:7]), 0600); err != nil {
		t.Fatal(err)
	}
	active := &ActiveDownload{
		Item:        item,
		PartialPath: partialPath,
		FinalPath:   finalPath,
	}
	if err := doDownload(context.Background(), item.URL, dir, item.ID, item.FileName, active); err != nil {
		t.Fatalf("doDownload returned error: %v", err)
	}
	if got := receivedRange.Load(); got != "bytes=7-" {
		t.Fatalf("Range header = %v, want bytes=7-", got)
	}
	got, err := os.ReadFile(finalPath)
	if err != nil {
		t.Fatalf("could not read completed file: %v", err)
	}
	if string(got) != content {
		t.Fatalf("downloaded contents = %q, want %q", got, content)
	}
	if _, err := os.Stat(partialPath); !os.IsNotExist(err) {
		t.Fatalf("partial file still exists after completion: %v", err)
	}
}

func TestPauseAndResumeDownloadUsesSavedPartialFile(t *testing.T) {
	downloadManager.ClearAll()
	t.Cleanup(ShutdownDownloads)

	const content = "download pauses and resumes"
	firstChunkSent := make(chan struct{}, 1)
	releaseFirstResponse := make(chan struct{})
	var releaseOnce sync.Once
	releaseHandler := func() { releaseOnce.Do(func() { close(releaseFirstResponse) }) }
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rangeHeader := r.Header.Get("Range")
		if rangeHeader == "" {
			w.Header().Set("ETag", `"version-1"`)
			w.Header().Set("Content-Length", fmt.Sprint(len(content)))
			_, _ = io.WriteString(w, content[:7])
			w.(http.Flusher).Flush()
			firstChunkSent <- struct{}{}
			<-releaseFirstResponse
			_, _ = io.WriteString(w, content[7:])
			return
		}
		if rangeHeader != "bytes=7-" {
			http.Error(w, "unexpected range", http.StatusBadRequest)
			return
		}
		if got := r.Header.Get("If-Range"); got != `"version-1"` {
			http.Error(w, "missing If-Range validator", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Range", fmt.Sprintf("bytes 7-%d/%d", len(content)-1, len(content)))
		w.Header().Set("Content-Length", fmt.Sprint(len(content)-7))
		w.WriteHeader(http.StatusPartialContent)
		_, _ = io.WriteString(w, content[7:])
	}))
	defer server.Close()
	defer releaseHandler()

	dir := t.TempDir()
	item := StartDownload(server.URL+"/file.bin", dir)
	select {
	case <-firstChunkSent:
	case <-time.After(5 * time.Second):
		t.Fatal("server did not send the first download chunk")
	}
	waitForDownloadState(t, func() bool {
		for _, stat := range GetDownloadStats() {
			if stat.ID == item.ID && stat.Received >= 7 {
				return true
			}
		}
		return false
	})

	if paused := ToggleDownloadPause(item.ID); !paused {
		t.Fatal("ToggleDownloadPause did not pause the active transfer")
	}
	waitForDownloadState(t, func() bool {
		for _, download := range downloadManager.GetAll() {
			if download.ID == item.ID {
				return download.Status == "paused"
			}
		}
		return false
	})
	releaseHandler()

	if paused := ToggleDownloadPause(item.ID); paused {
		t.Fatal("ToggleDownloadPause did not resume the paused transfer")
	}
	waitForDownloadState(t, func() bool {
		for _, download := range downloadManager.GetAll() {
			if download.ID == item.ID {
				return download.Status == "completed"
			}
		}
		return false
	})

	got, err := os.ReadFile(filepath.Join(dir, "file.bin"))
	if err != nil {
		t.Fatalf("could not read completed download: %v", err)
	}
	if string(got) != content {
		t.Fatalf("downloaded contents = %q, want %q", got, content)
	}
}

func waitForDownloadState(t *testing.T, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("timed out waiting for download state")
}
