package main

import "sync"

// ---------------------------------------------------
// PICTURE-IN-PICTURE HELPER
// Wraps the browser-native PiP API so SPARROW can trigger it
// from the toolbar instead of requiring a right-click on video
// ---------------------------------------------------

type PiPManager struct {
	mu     sync.Mutex
	Active map[int]bool // tabID -> pip currently active
}

var pipManager = &PiPManager{
	Active: make(map[int]bool),
}

func (p *PiPManager) SetActive(tabID int, active bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if active {
		p.Active[tabID] = true
	} else {
		delete(p.Active, tabID)
	}
}

func (p *PiPManager) IsActive(tabID int) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.Active[tabID]
}

// Injectable JS that finds the first video on the page and
// requests Picture-in-Picture for it
const pipRequestJS = `
(function() {
	var videos = document.querySelectorAll('video');
	if (videos.length === 0) {
		alert('No video found on this page');
		return;
	}
	var video = videos[0];
	if (document.pictureInPictureElement) {
		document.exitPictureInPicture();
	} else {
		video.requestPictureInPicture().catch(function(err) {
			console.log('PiP failed:', err);
		});
	}
})();
`
