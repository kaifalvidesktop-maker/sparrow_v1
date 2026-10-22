package main

import (
	"os"
	"strings"
)

// System media card (SMTC) off.
// Chromium honours only ONE --disable-features switch (the last one), and
// flashfix.go already sets one, so we MERGE into it instead of adding another.
// MediaSessionService stays ON so auto-PiP (mediaSession hook) keeps working.
func init() {
	const add = "GlobalMediaControls,HardwareMediaKeyHandling"
	const key = "WEBVIEW2_ADDITIONAL_BROWSER_ARGUMENTS"
	cur := strings.TrimSpace(os.Getenv(key))
	const p = "--disable-features="
	if i := strings.Index(cur, p); i >= 0 {
		j := i + len(p)
		cur = cur[:j] + add + "," + cur[j:]
	} else {
		cur = strings.TrimSpace(cur + " " + p + add)
	}
	os.Setenv(key, cur)
}

// Own video card (Chrome-style Document Picture-in-Picture):
// leave app while video plays -> floating always-on-top card with own buttons
// come back to app            -> card closes, video returns to the page.
const pipAutoJS = `
(function () {
  if (window.__sparrowPip) return;
  window.__sparrowPip = true;

  var cur = null, ph = null, pw = null, auto = false, oldStyle = null;

  function pick() {
    var vs = document.querySelectorAll('video'), i;
    for (i = 0; i < vs.length; i++) if (!vs[i].paused && !vs[i].ended && vs[i].readyState > 2) return vs[i];
    return null;
  }
  function fmt(t) {
    if (!isFinite(t)) return '0:00';
    var h = Math.floor(t / 3600), m = Math.floor(t % 3600 / 60), s = Math.floor(t % 60);
    return (h ? h + ':' + (m < 10 ? '0' : '') : '') + m + ':' + (s < 10 ? '0' : '') + s;
  }

  function restore() {
    if (!cur) return;
    try {
      if (ph && ph.parentNode) { ph.parentNode.insertBefore(cur, ph); ph.parentNode.removeChild(ph); }
      if (oldStyle === null) cur.removeAttribute('style'); else cur.setAttribute('style', oldStyle);
    } catch (e) {}
    cur = null; ph = null; pw = null; auto = false; oldStyle = null;
  }

  function build(d, v) {
    var st = d.createElement('style');
    st.textContent =
      'html,body{margin:0;height:100%;background:#000;overflow:hidden;font-family:Segoe UI,Arial,sans-serif;color:#fff}' +
      '#stage{position:relative;width:100%;height:100%}' +
      '#stage video{width:100%;height:100%;object-fit:contain;background:#000}' +
      '.bar{position:absolute;left:0;right:0;opacity:0;transition:opacity .2s;display:flex;align-items:center;gap:6px;padding:8px;box-sizing:border-box}' +
      '#stage:hover .bar{opacity:1}' +
      '#top{top:0;justify-content:space-between;background:linear-gradient(#000a,#0000)}' +
      '#bot{bottom:0;flex-wrap:wrap;background:linear-gradient(#0000,#000c)}' +
      'button{background:#fff2;border:0;color:#fff;border-radius:6px;min-width:30px;height:30px;font-size:14px;cursor:pointer;padding:0 8px}' +
      'button:hover{background:#fff5}' +
      'input[type=range]{accent-color:#fff;cursor:pointer}' +
      '#seek{flex:1 1 100%}#vol{width:70px}' +
      '#time{font-size:12px;margin-left:auto;white-space:nowrap}';
    d.head.appendChild(st);

    var stage = d.createElement('div'); stage.id = 'stage';
    var top = d.createElement('div'); top.className = 'bar'; top.id = 'top';
    var bot = d.createElement('div'); bot.className = 'bar'; bot.id = 'bot';

    function btn(txt, title, fn) {
      var b = d.createElement('button'); b.textContent = txt; b.title = title;
      b.onclick = function (e) { e.stopPropagation(); fn(b); };
      return b;
    }
    var back = btn('\u2197 Back to tab', 'Back to tab', function () { pw.close(); });
    var close = btn('\u2715', 'Close', function () { try { v.pause(); } catch (e) {} pw.close(); });
    top.appendChild(back); top.appendChild(close);

    var seek = d.createElement('input'); seek.type = 'range'; seek.id = 'seek'; seek.min = 0; seek.max = 1000; seek.value = 0;
    seek.oninput = function () { if (isFinite(v.duration)) v.currentTime = v.duration * seek.value / 1000; };

    var play = btn(v.paused ? '\u25B6' : '\u275A\u275A', 'Play/Pause', function () { v.paused ? v.play() : v.pause(); });
    var rw = btn('\u21BA10', 'Back 10s', function () { v.currentTime = Math.max(0, v.currentTime - 10); });
    var ff = btn('10\u21BB', 'Forward 10s', function () { v.currentTime = Math.min(v.duration || 1e9, v.currentTime + 10); });
    var mute = btn(v.muted ? '\uD83D\uDD07' : '\uD83D\uDD0A', 'Mute', function () { v.muted = !v.muted; });
    var vol = d.createElement('input'); vol.type = 'range'; vol.id = 'vol'; vol.min = 0; vol.max = 1; vol.step = 0.05; vol.value = v.volume;
    vol.oninput = function () { v.volume = parseFloat(vol.value); v.muted = v.volume === 0; };
    var speeds = [0.5, 0.75, 1, 1.25, 1.5, 2];
    var spd = btn(v.playbackRate + 'x', 'Speed', function (b) {
      var i = speeds.indexOf(v.playbackRate); v.playbackRate = speeds[(i + 1) % speeds.length]; b.textContent = v.playbackRate + 'x';
    });
    var time = d.createElement('span'); time.id = 'time'; time.textContent = '0:00 / 0:00';

    [seek, rw, play, ff, mute, vol, spd, time].forEach(function (e) { bot.appendChild(e); });

    function sync() {
      play.textContent = v.paused ? '\u25B6' : '\u275A\u275A';
      mute.textContent = (v.muted || v.volume === 0) ? '\uD83D\uDD07' : '\uD83D\uDD0A';
      if (isFinite(v.duration) && v.duration > 0) seek.value = Math.round(v.currentTime / v.duration * 1000);
      time.textContent = fmt(v.currentTime) + ' / ' + fmt(v.duration);
    }
    ['timeupdate', 'play', 'pause', 'volumechange', 'durationchange', 'ratechange'].forEach(function (ev) { v.addEventListener(ev, sync); });
    stage.ondblclick = function () { d.fullscreenElement ? d.exitFullscreen() : stage.requestFullscreen().catch(function () {}); };
    stage.onclick = function () { v.paused ? v.play() : v.pause(); };

    stage.appendChild(v); stage.appendChild(top); stage.appendChild(bot);
    d.body.appendChild(stage);
    sync();
    d.addEventListener('keydown', function (e) {
      if (e.code === 'Space') { e.preventDefault(); v.paused ? v.play() : v.pause(); }
      else if (e.code === 'ArrowLeft') v.currentTime -= 5;
      else if (e.code === 'ArrowRight') v.currentTime += 5;
      else if (e.key === 'm') v.muted = !v.muted;
    });
  }

  async function openCard(isAuto) {
    if (pw || cur) return;
    var v = pick();
    if (!v) return;
    if (!window.documentPictureInPicture) {
      if (document.pictureInPictureEnabled && !document.pictureInPictureElement) v.requestPictureInPicture().catch(function () {});
      return;
    }
    try {
      var r = v.getBoundingClientRect();
      var w = Math.max(360, Math.min(640, Math.round(r.width) || 480));
      var h = Math.max(220, Math.min(400, Math.round(r.height) || 270));
      pw = await documentPictureInPicture.requestWindow({ width: w, height: h });
    } catch (e) {
      pw = null;
      if (document.pictureInPictureEnabled && !document.pictureInPictureElement) v.requestPictureInPicture().catch(function () {});
      return;
    }
    cur = v; auto = !!isAuto;
    oldStyle = v.getAttribute('style');
    ph = document.createComment('sparrow-pip');
    v.parentNode.insertBefore(ph, v);
    build(pw.document, v);
    pw.addEventListener('pagehide', restore);
  }

  function closeCard() {
    if (pw && auto) { try { pw.close(); } catch (e) {} }
  }

  window.addEventListener('blur', function () {
    setTimeout(function () { if (!document.hasFocus()) openCard(true); }, 150);
  });
  window.addEventListener('focus', closeCard);
  document.addEventListener('visibilitychange', function () { document.hidden ? openCard(true) : closeCard(); });

  try { navigator.mediaSession.setActionHandler('enterpictureinpicture', function () { openCard(true); }); } catch (e) {}
})();
`
