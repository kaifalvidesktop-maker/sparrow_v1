package main

import "os"

// Disable the Windows system media card (SMTC / global media controls)
// that WebView2 shows when you leave the app while a video plays.
// init() runs before main(), so this is set before webview.New().
func init() {
	const flags = "--disable-features=GlobalMediaControls,HardwareMediaKeyHandling,MediaSessionService"
	if cur := os.Getenv("WEBVIEW2_ADDITIONAL_BROWSER_ARGUMENTS"); cur != "" {
		os.Setenv("WEBVIEW2_ADDITIONAL_BROWSER_ARGUMENTS", cur+" "+flags)
	} else {
		os.Setenv("WEBVIEW2_ADDITIONAL_BROWSER_ARGUMENTS", flags)
	}
}

// Own video card (Picture-in-Picture):
// app loses focus + video playing -> PiP opens
// app gets focus again            -> PiP closes
const pipAutoJS = `
(function () {
  if (window.__sparrowPipAuto) return;
  window.__sparrowPipAuto = true;

  var autoOpened = false;

  function playingVideo() {
    var vs = document.querySelectorAll('video');
    for (var i = 0; i < vs.length; i++) {
      var v = vs[i];
      if (!v.paused && !v.ended && v.readyState > 2 && !v.disablePictureInPicture) return v;
    }
    return null;
  }

  function openPip() {
    if (document.pictureInPictureElement) return;
    var v = playingVideo();
    if (!v) return;
    v.requestPictureInPicture().then(function () { autoOpened = true; }).catch(function () {});
  }

  function closePip() {
    if (autoOpened && document.pictureInPictureElement) {
      document.exitPictureInPicture().catch(function () {});
    }
    autoOpened = false;
  }

  window.addEventListener('blur', function () { setTimeout(function () { if (!document.hasFocus()) openPip(); }, 150); });
  window.addEventListener('focus', closePip);
  document.addEventListener('visibilitychange', function () {
    if (document.hidden) openPip(); else closePip();
  });

  // Chromium auto-PiP hook (works without a user gesture when supported)
  try {
    navigator.mediaSession.setActionHandler('enterpictureinpicture', function () { openPip(); });
  } catch (e) {}

  // Clear media-session metadata so no system card gets content
  try {
    navigator.mediaSession.metadata = null;
  } catch (e) {}
})();
`
