package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"math"
	"mime"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ---------------------------------------------------
// REAL DOWNLOAD ENGINE (actually saves files to disk)
// ---------------------------------------------------

type ActiveDownload struct {
	mu              sync.Mutex
	Item            *DownloadItem
	URL             string
	SaveDir         string
	Paused          bool
	Cancelled       bool
	Running         bool
	Queued          bool
	PauseInFlight   bool
	ResumeRequested bool
	PartialPath     string
	FinalPath       string
	Validator       string
	ValidatorHeader string
	cancel          context.CancelFunc
}

const maxConcurrentDownloads = 3

var activeDownloads = struct {
	mu    sync.Mutex
	Items map[int]*ActiveDownload
}{Items: make(map[int]*ActiveDownload)}

var activeWG sync.WaitGroup

var downloadQueue = struct {
	mu       sync.Mutex
	pending  []int
	running  int
	stopping bool
}{}

// StartDownload adds a transfer to the FIFO queue. At most three transfers run at once.
func StartDownload(rawURL string, saveDir string) *DownloadItem {
	item := downloadManager.Start(rawURL)
	snapshot := *item
	downloadManager.SetStatus(item.ID, "queued")
	active := &ActiveDownload{
		Item:    item,
		URL:     rawURL,
		SaveDir: resolveSaveDir(saveDir),
		Queued:  true,
	}
	activeDownloads.mu.Lock()
	activeDownloads.Items[item.ID] = active
	activeDownloads.mu.Unlock()
	updateQueueStat(item.ID, 0, 0, true)
	enqueueDownload(item.ID)

	return &snapshot
}

func enqueueDownload(id int) {
	downloadQueue.mu.Lock()
	defer downloadQueue.mu.Unlock()
	if downloadQueue.stopping {
		return
	}
	downloadQueue.pending = append(downloadQueue.pending, id)
	startQueuedDownloadsLocked()
}

func startQueuedDownloadsLocked() {
	for !downloadQueue.stopping && downloadQueue.running < maxConcurrentDownloads && len(downloadQueue.pending) > 0 {
		id := downloadQueue.pending[0]
		downloadQueue.pending = downloadQueue.pending[1:]

		activeDownloads.mu.Lock()
		active := activeDownloads.Items[id]
		activeDownloads.mu.Unlock()
		if active == nil {
			continue
		}

		active.mu.Lock()
		if active.Cancelled || active.Paused || active.Running {
			active.Queued = false
			active.mu.Unlock()
			continue
		}
		ctx, cancel := context.WithCancel(context.Background())
		active.cancel = cancel
		active.Running = true
		active.Queued = false
		active.mu.Unlock()

		downloadManager.SetStatus(id, "downloading")
		updateRunningDownloadStat(id, active)
		downloadQueue.running++
		activeWG.Add(1)
		go runDownload(ctx, active)
	}
}

func runDownload(ctx context.Context, active *ActiveDownload) {
	defer activeWG.Done()
	err := doDownload(ctx, active.URL, active.SaveDir, active.Item.ID, active.Item.FileName, active)

	active.mu.Lock()
	cancelled := active.Cancelled && err != nil
	paused := err != nil && active.PauseInFlight && !cancelled
	resume := paused && active.ResumeRequested
	active.Running = false
	active.Queued = false
	active.PauseInFlight = false
	active.ResumeRequested = false
	active.cancel = nil
	if resume && !cancelled {
		active.Paused = false
		active.Queued = true
	} else if paused {
		active.Paused = true
	}
	partialPath := active.PartialPath
	active.mu.Unlock()

	switch {
	case cancelled:
		if partialPath != "" {
			_ = os.Remove(partialPath)
		}
		downloadManager.Cancel(active.Item.ID)
		updateStat(active.Item.ID, 0, 0, 0, false)
		cleanupActiveDownload(active.Item.ID)
	case paused:
		updateDownloadPausedStat(active.Item.ID)
		if resume {
			downloadManager.SetStatus(active.Item.ID, "queued")
			updateQueueStat(active.Item.ID, 0, 0, true)
		} else {
			downloadManager.SetStatus(active.Item.ID, "paused")
		}
	case err != nil:
		if partialPath != "" {
			_ = os.Remove(partialPath)
		}
		downloadManager.Fail(active.Item.ID)
		cleanupActiveDownload(active.Item.ID)
	default:
		downloadManager.Complete(active.Item.ID)
		cleanupActiveDownload(active.Item.ID)
	}

	downloadQueue.mu.Lock()
	downloadQueue.running--
	if resume && !cancelled && !downloadQueue.stopping {
		downloadQueue.pending = append(downloadQueue.pending, active.Item.ID)
	}
	startQueuedDownloadsLocked()
	downloadQueue.mu.Unlock()
}

// resolveSaveDir turns the saved setting ("Downloads" by default) into a real folder
func resolveSaveDir(saveDir string) string {
	saveDir = strings.TrimSpace(saveDir)
	if filepath.IsAbs(saveDir) {
		return saveDir
	}
	if dir, err := userDownloadsDirectory(); err == nil {
		if saveDir == "" || strings.EqualFold(saveDir, "Downloads") {
			return dir
		}
		if home, herr := os.UserHomeDir(); herr == nil {
			return filepath.Join(home, saveDir)
		}
		return dir
	}
	if saveDir == "" {
		return "Downloads"
	}
	return saveDir
}

func doDownload(ctx context.Context, rawURL string, saveDir string, id int, fallbackName string, active *ActiveDownload) (err error) {
	client := &http.Client{Transport: &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		ResponseHeaderTimeout: 30 * time.Second,
		TLSHandshakeTimeout:   15 * time.Second,
	}}
	defer client.CloseIdleConnections()
	active.mu.Lock()
	partialPath, finalPath := active.PartialPath, active.FinalPath
	validator, validatorHeader := active.Validator, active.ValidatorHeader
	active.mu.Unlock()
	var received int64
	if partialPath != "" {
		if info, statErr := os.Stat(partialPath); statErr == nil {
			received = info.Size()
		} else if !os.IsNotExist(statErr) {
			return statErr
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0 Safari/537.36 SPARROW")
	req.Header.Set("Accept", "*/*")
	if received > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", received))
		if validator != "" && validatorHeader != "" {
			req.Header.Set("If-Range", validator)
		}
	}
	if u, perr := url.Parse(rawURL); perr == nil && u.Host != "" {
		req.Header.Set("Referer", u.Scheme+"://"+u.Host+"/")
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusRequestedRangeNotSatisfiable && received > 0 {
		if total, ok := parseUnsatisfiedContentRange(resp.Header.Get("Content-Range")); ok && total == received {
			return finishPartialDownload(id, partialPath, finalPath, received, active)
		}
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return errNotOK(resp.StatusCode)
	}
	if received > 0 && resp.StatusCode == http.StatusOK {
		// The server ignored Range; safely restart instead of appending duplicate bytes.
		received = 0
	}
	if received > 0 && resp.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("server returned HTTP %d for a resumed download", resp.StatusCode)
	}
	if resp.StatusCode == http.StatusPartialContent {
		rangeStart, _, _, ok := parseContentRange(resp.Header.Get("Content-Range"))
		if !ok || rangeStart != received {
			return errors.New("server returned an invalid Content-Range for a resumed download")
		}
	}
	if received == 0 {
		active.mu.Lock()
		active.Validator, active.ValidatorHeader = downloadValidator(resp)
		active.mu.Unlock()
	}

	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return err
	}

	if partialPath == "" {
		name := downloadNameFromResponse(resp, fallbackName)
		partialPath, finalPath, err = createPartialDownloadFile(saveDir, sanitizeFileName(name))
		if err != nil {
			return err
		}
		active.mu.Lock()
		active.PartialPath = partialPath
		active.FinalPath = finalPath
		active.mu.Unlock()
		downloadManager.SetFile(id, filepath.Base(finalPath), finalPath)
	}

	flags := os.O_WRONLY | os.O_CREATE
	if received == 0 {
		flags |= os.O_TRUNC
	} else {
		flags |= os.O_APPEND
	}
	out, err := os.OpenFile(partialPath, flags, 0644)
	if err != nil {
		return err
	}
	closed := false
	defer func() {
		if !closed {
			if closeErr := out.Close(); err == nil && closeErr != nil {
				err = closeErr
			}
		}
	}()

	total := resp.ContentLength
	if resp.StatusCode == http.StatusPartialContent {
		if _, _, rangeTotal, ok := parseContentRange(resp.Header.Get("Content-Range")); ok && rangeTotal >= 0 {
			total = rangeTotal
		} else if total >= 0 {
			total += received
		}
	}
	written := received
	buf := make([]byte, 32*1024)
	lastPercent := -1
	lastT, lastW := time.Now(), int64(0)
	speed := int64(0)

	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := out.Write(buf[:n]); writeErr != nil {
				return writeErr
			}
			written += int64(n)
			updateStat(id, written, total, speed, false)
			if total > 0 {
				percent := int((written * 100) / total)
				if percent != lastPercent {
					lastPercent = percent
					downloadManager.UpdateProgress(id, percent)
				}
			} else {
				// unknown size: show smooth progress instead of a stuck 0%
				percent := int(99 * (1 - math.Exp(-float64(written)/float64(20<<20))))
				if percent != lastPercent {
					lastPercent = percent
					downloadManager.UpdateProgress(id, percent)
				}
			}
		}
		if d := time.Since(lastT); d >= 500*time.Millisecond {
			speed = int64(float64(written-lastW) / d.Seconds())
			lastT, lastW = time.Now(), written
			updateStat(id, written, total, speed, false)
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return readErr
		}
	}

	if err = out.Close(); err != nil {
		return err
	}
	closed = true
	if err = os.Rename(partialPath, finalPath); err != nil {
		return err
	}
	active.mu.Lock()
	active.PartialPath = ""
	active.mu.Unlock()
	downloadManager.UpdateProgress(id, 100)
	updateStat(id, written, total, 0, false)
	return nil
}

func downloadNameFromResponse(resp *http.Response, fallback string) string {
	name := fallback
	if cd := resp.Header.Get("Content-Disposition"); cd != "" {
		if _, params, err := mime.ParseMediaType(cd); err == nil {
			if fn := strings.TrimSpace(params["filename"]); fn != "" {
				name = path.Base(strings.ReplaceAll(fn, "\\", "/"))
			}
		}
	}
	if filepath.Ext(name) == "" {
		if resp.Request != nil && resp.Request.URL != nil {
			if redirected := fileNameFromURL(resp.Request.URL.String()); redirected != "unknown_file" && filepath.Ext(redirected) != "" {
				name = redirected
			}
		}
		if filepath.Ext(name) == "" {
			if contentType, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type")); err == nil && contentType != "text/html" {
				if extensions, _ := mime.ExtensionsByType(contentType); len(extensions) > 0 {
					name += extensions[0]
				}
			}
		}
	}
	return name
}

func downloadValidator(resp *http.Response) (string, string) {
	if etag := strings.TrimSpace(resp.Header.Get("ETag")); etag != "" && !strings.HasPrefix(strings.ToUpper(etag), "W/") {
		return etag, "If-Range"
	}
	if modified := strings.TrimSpace(resp.Header.Get("Last-Modified")); modified != "" {
		return modified, "If-Range"
	}
	return "", ""
}

func updateRunningDownloadStat(id int, active *ActiveDownload) {
	received := int64(0)
	active.mu.Lock()
	partialPath := active.PartialPath
	active.mu.Unlock()
	if partialPath != "" {
		if info, err := os.Stat(partialPath); err == nil {
			received = info.Size()
		}
	}
	dlStats.mu.Lock()
	stat := dlStats.M[id]
	total := int64(0)
	if stat != nil {
		total = stat.Total
	}
	dlStats.M[id] = &DownloadStat{ID: id, Received: received, Total: total}
	dlStats.mu.Unlock()
}

func createPartialDownloadFile(dir, name string) (string, string, error) {
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)
	for i := 0; i < 10000; i++ {
		candidate := name
		if i > 0 {
			candidate = base + " (" + strconv.Itoa(i) + ")" + ext
		}
		finalPath := filepath.Join(dir, candidate)
		if _, err := os.Stat(finalPath); err == nil {
			continue
		} else if !os.IsNotExist(err) {
			return "", "", err
		}
		partialPath := finalPath + ".sparrow-part"
		file, err := os.OpenFile(partialPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
		if err == nil {
			if closeErr := file.Close(); closeErr != nil {
				os.Remove(partialPath)
				return "", "", closeErr
			}
			return partialPath, finalPath, nil
		}
		if !os.IsExist(err) {
			return "", "", err
		}
	}
	return "", "", os.ErrExist
}

func parseContentRange(value string) (start, end, total int64, ok bool) {
	if !strings.HasPrefix(value, "bytes ") {
		return 0, 0, 0, false
	}
	parts := strings.SplitN(strings.TrimPrefix(value, "bytes "), "/", 2)
	if len(parts) != 2 {
		return 0, 0, 0, false
	}
	bounds := strings.SplitN(parts[0], "-", 2)
	if len(bounds) != 2 {
		return 0, 0, 0, false
	}
	start, err := strconv.ParseInt(bounds[0], 10, 64)
	if err != nil {
		return 0, 0, 0, false
	}
	end, err = strconv.ParseInt(bounds[1], 10, 64)
	if err != nil || end < start {
		return 0, 0, 0, false
	}
	if parts[1] == "*" {
		return start, end, -1, true
	}
	total, err = strconv.ParseInt(parts[1], 10, 64)
	if err != nil || total <= end {
		return 0, 0, 0, false
	}
	return start, end, total, true
}

func parseUnsatisfiedContentRange(value string) (int64, bool) {
	if !strings.HasPrefix(value, "bytes */") {
		return 0, false
	}
	total, err := strconv.ParseInt(strings.TrimPrefix(value, "bytes */"), 10, 64)
	return total, err == nil && total >= 0
}

func finishPartialDownload(id int, partialPath, finalPath string, received int64, active *ActiveDownload) error {
	if partialPath == "" || finalPath == "" {
		return errors.New("download has no partial file to finish")
	}
	if err := os.Rename(partialPath, finalPath); err != nil {
		return err
	}
	downloadManager.SetFile(id, filepath.Base(finalPath), finalPath)
	downloadManager.UpdateProgress(id, 100)
	updateStat(id, received, received, 0, false)
	active.mu.Lock()
	active.PartialPath = ""
	active.mu.Unlock()
	return nil
}

// createUniqueFile creates name, or name (1).ext, name (2).ext ... if it exists
func createUniqueFile(dir string, name string) (string, *os.File, error) {
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)
	candidate := name
	for i := 1; i < 10000; i++ {
		full := filepath.Join(dir, candidate)
		f, err := os.OpenFile(full, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
		if err == nil {
			return full, f, nil
		}
		if !os.IsExist(err) {
			return "", nil, err
		}
		candidate = base + " (" + strconv.Itoa(i) + ")" + ext
	}
	return "", nil, os.ErrExist
}

func cleanupActiveDownload(id int) {
	activeDownloads.mu.Lock()
	delete(activeDownloads.Items, id)
	activeDownloads.mu.Unlock()
}

func removePendingDownloadLocked(id int) {
	pending := downloadQueue.pending[:0]
	for _, pendingID := range downloadQueue.pending {
		if pendingID != id {
			pending = append(pending, pendingID)
		}
	}
	downloadQueue.pending = pending
}

// CancelActiveDownload cancels queued or active transfers and removes partial files.
func CancelActiveDownload(id int) {
	downloadQueue.mu.Lock()
	activeDownloads.mu.Lock()
	active, ok := activeDownloads.Items[id]
	activeDownloads.mu.Unlock()

	if !ok {
		downloadManager.Cancel(id)
		downloadQueue.mu.Unlock()
		return
	}

	active.mu.Lock()
	active.Cancelled = true
	active.Paused = false
	active.ResumeRequested = false
	cancel := active.cancel
	wasRunning := active.Running
	partialPath := active.PartialPath
	if !wasRunning {
		active.Queued = false
		removePendingDownloadLocked(id)
	}
	active.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	if !wasRunning {
		if partialPath != "" {
			_ = os.Remove(partialPath)
		}
		downloadManager.Cancel(id)
		cleanupActiveDownload(id)
	}
	startQueuedDownloadsLocked()
	downloadQueue.mu.Unlock()
}

func updateDownloadPausedStat(id int) {
	dlStats.mu.Lock()
	stat := dlStats.M[id]
	if stat == nil {
		stat = &DownloadStat{ID: id}
		dlStats.M[id] = stat
	}
	stat.Speed = 0
	stat.Paused = true
	stat.Queued = false
	dlStats.mu.Unlock()
}

// ShutdownDownloads is called when the app closes: stops every transfer,
// removes partial files and clears the download list.
func ShutdownDownloads() {
	downloadQueue.mu.Lock()
	downloadQueue.stopping = true
	downloadQueue.pending = nil
	activeDownloads.mu.Lock()
	queuedIDs := make([]int, 0)
	for id, active := range activeDownloads.Items {
		active.mu.Lock()
		active.Cancelled = true
		active.Paused = false
		cancel := active.cancel
		if !active.Running {
			active.Queued = false
			queuedIDs = append(queuedIDs, id)
		}
		active.mu.Unlock()
		if cancel != nil {
			cancel()
		}
	}
	activeDownloads.mu.Unlock()
	downloadQueue.mu.Unlock()

	for _, id := range queuedIDs {
		activeDownloads.mu.Lock()
		active := activeDownloads.Items[id]
		activeDownloads.mu.Unlock()
		if active != nil {
			active.mu.Lock()
			partialPath := active.PartialPath
			active.mu.Unlock()
			if partialPath != "" {
				_ = os.Remove(partialPath)
			}
			cleanupActiveDownload(id)
		}
	}

	done := make(chan struct{})
	go func() { activeWG.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
	}
	downloadManager.ClearAll()
	dlStats.mu.Lock()
	dlStats.M = map[int]*DownloadStat{}
	dlStats.mu.Unlock()
	downloadQueue.mu.Lock()
	downloadQueue.running = 0
	downloadQueue.pending = nil
	downloadQueue.stopping = false
	downloadQueue.mu.Unlock()
}

func sanitizeFileName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" || name == "." || name == ".." {
		name = "download"
	}
	replacer := strings.NewReplacer(
		"<", "_", ">", "_", ":", "_", "\"", "_",
		"/", "_", "\\", "_", "|", "_", "?", "_", "*", "_",
	)
	return replacer.Replace(name)
}

type downloadHTTPError struct {
	code int
}

func (e *downloadHTTPError) Error() string {
	return "download failed with status code"
}

func errNotOK(code int) error {
	return &downloadHTTPError{code: code}
}

// IsDownloadableURL checks file extensions that should trigger a download
// instead of opening in the browser (basic heuristic)
func IsDownloadableURL(rawURL string) bool {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return false
	}
	lower := strings.ToLower(parsed.Path)
	downloadExts := []string{
		".zip", ".rar", ".7z", ".exe", ".msi", ".msix", ".msixbundle", ".appx", ".appxbundle",
		".dmg", ".pkg", ".deb", ".rpm", ".jar", ".bin", ".img", ".iso", ".apk", ".ipa",
		".tar", ".gz", ".tgz", ".bz2", ".xz", ".zst", ".cab", ".whl", ".torrent",
		".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", ".csv", ".rtf", ".odt", ".ods", ".odp", ".epub",
		".mp3", ".wav", ".flac", ".m4a", ".aac", ".mp4", ".mkv", ".avi", ".mov", ".wmv", ".flv",
	}
	for _, ext := range downloadExts {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	return false
}

// SaveDownloadData saves bytes produced inside a page (blob:/data: downloads).
func SaveDownloadData(name string, b64 string, saveDir string) *DownloadItem {
	item := downloadManager.Start("local://" + name)
	snapshot := *item
	data, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		downloadManager.Fail(item.ID)
		return &snapshot
	}
	dir := resolveSaveDir(saveDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		downloadManager.Fail(item.ID)
		return &snapshot
	}
	fullPath, out, err := createUniqueFile(dir, sanitizeFileName(name))
	if err != nil {
		downloadManager.Fail(item.ID)
		return &snapshot
	}
	downloadManager.SetFile(item.ID, filepath.Base(fullPath), fullPath)
	_, werr := out.Write(data)
	cerr := out.Close()
	if werr != nil || cerr != nil {
		os.Remove(fullPath)
		downloadManager.Fail(item.ID)
		return &snapshot
	}
	updateStat(item.ID, int64(len(data)), int64(len(data)), 0, false)
	downloadManager.Complete(item.ID)
	return &snapshot
}

// ---------------------------------------------------
// LIVE STATS, PAUSE/RESUME, OPEN FILE
// ---------------------------------------------------

type DownloadStat struct {
	ID       int   `json:"id"`
	Received int64 `json:"received"`
	Total    int64 `json:"total"`
	Speed    int64 `json:"speed"`
	Paused   bool  `json:"paused"`
	Queued   bool  `json:"queued"`
}

var dlStats = struct {
	mu sync.Mutex
	M  map[int]*DownloadStat
}{M: map[int]*DownloadStat{}}

func updateStat(id int, received, total, speed int64, paused bool) {
	if total < 0 {
		total = 0
	}
	dlStats.mu.Lock()
	dlStats.M[id] = &DownloadStat{ID: id, Received: received, Total: total, Speed: speed, Paused: paused}
	dlStats.mu.Unlock()
}

func updateQueueStat(id int, received, total int64, queued bool) {
	dlStats.mu.Lock()
	dlStats.M[id] = &DownloadStat{ID: id, Received: received, Total: total, Queued: queued}
	dlStats.mu.Unlock()
}

func GetDownloadStats() []DownloadStat {
	dlStats.mu.Lock()
	defer dlStats.mu.Unlock()
	out := make([]DownloadStat, 0, len(dlStats.M))
	for _, st := range dlStats.M {
		out = append(out, *st)
	}
	return out
}

// ToggleDownloadPause pauses/resumes a queued or active download.
func ToggleDownloadPause(id int) bool {
	downloadQueue.mu.Lock()
	activeDownloads.mu.Lock()
	active, ok := activeDownloads.Items[id]
	activeDownloads.mu.Unlock()
	if !ok {
		downloadQueue.mu.Unlock()
		return false
	}

	active.mu.Lock()
	if active.Cancelled {
		active.mu.Unlock()
		downloadQueue.mu.Unlock()
		return false
	}

	if !active.Paused {
		active.Paused = true
		if active.Running {
			active.PauseInFlight = true
			active.ResumeRequested = false
			if active.cancel != nil {
				active.cancel()
			}
		} else {
			active.Queued = false
			removePendingDownloadLocked(id)
			downloadManager.SetStatus(id, "paused")
			updateDownloadPausedStat(id)
		}
	} else {
		active.Paused = false
		if active.PauseInFlight {
			active.ResumeRequested = true
		} else if !active.Running {
			active.Queued = true
			downloadManager.SetStatus(id, "queued")
			updateQueueStat(id, 0, 0, true)
			downloadQueue.pending = append(downloadQueue.pending, id)
		}
	}
	paused := active.Paused
	active.mu.Unlock()
	startQueuedDownloadsLocked()
	downloadQueue.mu.Unlock()
	return paused
}

// downloadPathByID only returns paths of completed downloads Sparrow itself saved,
// so pages can never ask the app to open arbitrary files.
func downloadPathByID(id int) string {
	for _, it := range downloadManager.GetAll() {
		if it.ID == id && it.Status == "completed" && it.Path != "" {
			return it.Path
		}
	}
	return ""
}

func OpenDownloadedFile(id int) {
	if p := downloadPathByID(id); p != "" {
		openWithSystem(p)
	}
}

func RevealDownloadedFile(id int) {
	p := downloadPathByID(id)
	if p == "" {
		return
	}
	switch runtime.GOOS {
	case "windows":
		exec.Command("explorer", "/select,"+p).Start()
	case "darwin":
		exec.Command("open", "-R", p).Start()
	default:
		exec.Command("xdg-open", filepath.Dir(p)).Start()
	}
}

func OpenFolder(dir string) {
	os.MkdirAll(dir, 0755)
	openWithSystem(dir)
}

func openWithSystem(p string) {
	switch runtime.GOOS {
	case "windows":
		exec.Command("rundll32", "url.dll,FileProtocolHandler", p).Start()
	case "darwin":
		exec.Command("open", p).Start()
	default:
		exec.Command("xdg-open", p).Start()
	}
}
