package main

import (
	"strconv"
	"sync"
)

type ReaderModeManager struct {
	mu     sync.Mutex
	Active map[int]bool
}

var readerMode = &ReaderModeManager{
	Active: make(map[int]bool),
}

func (r *ReaderModeManager) Toggle(tabID int) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.Active[tabID] = !r.Active[tabID]
	return r.Active[tabID]
}

func (r *ReaderModeManager) IsActive(tabID int) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.Active[tabID]
}

func (r *ReaderModeManager) Clear(tabID int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.Active, tabID)
}

func BuildReaderModeJS(enabled bool) string {
	return `;(function () {
  "use strict";
  var enabled = ` + strconv.FormatBool(enabled) + `;
  var STYLE_ID = "sparrow-reader-style";
  var PANEL_ID = "sparrow-reader-panel";
  var CONTENT_ATTR = "data-sparrow-reader-content";
  var current = document.getElementById(STYLE_ID);
  var panel = document.getElementById(PANEL_ID);
  var state = window.__sparrowReaderState;

  function restoreProperty(element, name, saved) {
    if (!element) return;
    if (saved.value) element.style.setProperty(name, saved.value, saved.priority);
    else element.style.removeProperty(name);
  }

  function restore() {
    if (state) {
      state.hidden.forEach(function (item) {
        restoreProperty(item.element, "display", item.display);
      });
      state.content.forEach(function (item) {
        restoreProperty(item.element, "font-family", item.fontFamily);
        restoreProperty(item.element, "font-size", item.fontSize);
        restoreProperty(item.element, "line-height", item.lineHeight);
        restoreProperty(item.element, "color", item.color);
        restoreProperty(item.element, "background-color", item.backgroundColor);
        restoreProperty(item.element, "max-width", item.maxWidth);
        restoreProperty(item.element, "margin-left", item.marginLeft);
        restoreProperty(item.element, "margin-right", item.marginRight);
        restoreProperty(item.element, "padding", item.padding);
        restoreProperty(item.element, "--sparrow-reader-font-size", item.readerFontSize);
        restoreProperty(item.element, "--sparrow-reader-color", item.readerColor);
        restoreProperty(item.element, "--sparrow-reader-link", item.readerLink);
      });
      state.marked.forEach(function (element) {
        element.removeAttribute(CONTENT_ATTR);
      });
      if (typeof state.scrollY === "number") window.scrollTo(0, state.scrollY);
    }
    if (current) current.remove();
    if (panel) panel.remove();
    delete window.__sparrowReaderState;
  }

  if (!enabled) {
    restore();
    return;
  }
  if (current && panel && state) return;
  if (current || panel || state) restore();
  if (!document.body) return;

  function score(element) {
    var textLength = (element.innerText || "").trim().length;
    var paragraphs = element.querySelectorAll("p").length;
    var headings = element.querySelectorAll("h1,h2,h3").length;
    return textLength + paragraphs * 180 + headings * 80;
  }

  var candidates = Array.prototype.slice.call(
    document.querySelectorAll("article, main, [role='main'], .article, .post-content, .entry-content")
  ).filter(function (element) {
    return !element.closest("#__sparrow_chrome") && score(element) >= 300;
  });
  var content = null;
  candidates.forEach(function (candidate) {
    if (!content || score(candidate) > score(content)) content = candidate;
  });

  if (!content) {
    var blocks = Array.prototype.slice.call(document.querySelectorAll("body div, body section"))
      .filter(function (element) {
        return !element.closest("#__sparrow_chrome") &&
          element.querySelectorAll("p").length >= 2 &&
          (element.innerText || "").trim().length >= 500;
      });
    blocks.forEach(function (candidate) {
      if (!content || score(candidate) > score(content)) content = candidate;
    });
  }
  if (!content) {
    content = document.body;
  }

  var marked = [];
  var scrollY = window.scrollY;
  var contentRoot = content;
  while (contentRoot.parentElement && contentRoot.parentElement !== document.body) {
    contentRoot = contentRoot.parentElement;
  }
  var hidden = [];
  if (contentRoot !== document.body) {
    Array.prototype.forEach.call(document.body.children, function (element) {
      if (element === contentRoot || element.id === "__sparrow_chrome" ||
          element.id === PANEL_ID || element.tagName === "SCRIPT" || element.tagName === "STYLE") return;
      hidden.push({
        element: element,
        display: {
          value: element.style.getPropertyValue("display"),
          priority: element.style.getPropertyPriority("display")
        }
      });
      element.style.setProperty("display", "none", "important");
    });
  }

  var contentElements = [content].concat(Array.prototype.slice.call(content.querySelectorAll("*")));
  contentElements.forEach(function (element) {
    if (element.closest("#__sparrow_chrome, #" + PANEL_ID)) return;
    marked.push(element);
    element.setAttribute(CONTENT_ATTR, "");
  });

  function savedStyle(element, name) {
    return {
      value: element.style.getPropertyValue(name),
      priority: element.style.getPropertyPriority(name)
    };
  }
  var contentStyles = [{
    element: content,
    fontFamily: savedStyle(content, "font-family"),
    fontSize: savedStyle(content, "font-size"),
    lineHeight: savedStyle(content, "line-height"),
    color: savedStyle(content, "color"),
    backgroundColor: savedStyle(content, "background-color"),
    maxWidth: savedStyle(content, "max-width"),
    marginLeft: savedStyle(content, "margin-left"),
    marginRight: savedStyle(content, "margin-right"),
    padding: savedStyle(content, "padding"),
    readerFontSize: savedStyle(content, "--sparrow-reader-font-size"),
    readerColor: savedStyle(content, "--sparrow-reader-color"),
    readerLink: savedStyle(content, "--sparrow-reader-link")
  }];
  window.__sparrowReaderState = {
    hidden: hidden,
    content: contentStyles,
    marked: marked,
    scrollY: scrollY
  };

  var style = document.createElement("style");
  style.id = STYLE_ID;
  style.textContent =
    "[" + CONTENT_ATTR + "] { font-family: Georgia, 'Times New Roman', serif !important; color: var(--sparrow-reader-color, #25231f) !important; }" +
    "[" + CONTENT_ATTR + "] p, [" + CONTENT_ATTR + "] li { font-size: var(--sparrow-reader-font-size, 20px) !important; line-height: 1.75 !important; }" +
    "[" + CONTENT_ATTR + "] h1, [" + CONTENT_ATTR + "] h2, [" + CONTENT_ATTR + "] h3 { line-height: 1.3 !important; margin-top: 1.5em !important; }" +
    "[" + CONTENT_ATTR + "] img, [" + CONTENT_ATTR + "] video { max-width: 100% !important; height: auto !important; }" +
    "[" + CONTENT_ATTR + "] a { color: var(--sparrow-reader-link, #245bb5) !important; }" +
    "[" + CONTENT_ATTR + "] [data-ad], [" + CONTENT_ATTR + "] [class*='share'], [" + CONTENT_ATTR + "] [class*='social'] { display: none !important; }" +
    "#"+ PANEL_ID + " { position: fixed !important; top: 132px !important; right: 20px !important; z-index: 2147483646 !important; display: flex !important; align-items: center !important; gap: 8px !important; padding: 8px !important; border: 1px solid rgba(151,154,255,.28) !important; border-radius: 12px !important; background: rgba(21,22,27,.96) !important; color: #e8e8e8 !important; box-shadow: 0 6px 30px rgba(85,96,255,.32) !important; font: 13px 'Segoe UI', sans-serif !important; }" +
    "#"+ PANEL_ID + " button { display: inline-flex !important; align-items: center !important; justify-content: center !important; min-width: 30px !important; height: 30px !important; padding: 0 8px !important; border: 1px solid rgba(151,154,255,.28) !important; border-radius: 7px !important; background: rgba(25,26,35,.9) !important; color: #eae9ff !important; font: inherit !important; cursor: pointer !important; }" +
    "#"+ PANEL_ID + " button:hover { background: rgba(120,96,255,.16) !important; }";
  (document.head || document.documentElement).appendChild(style);

  var fontSize = 20;
  var themes = [
    { name: "Dark", background: "#0b0b0d", color: "#e8e8e8", link: "#60a5fa" },
    { name: "Light", background: "#ffffff", color: "#25231f", link: "#245bb5" },
    { name: "Sepia", background: "#f4ecd8", color: "#433b2f", link: "#805b24" }
  ];
  var themeIndex = 0;
  content.style.setProperty("font-family", "Georgia, 'Times New Roman', serif", "important");
  content.style.setProperty("font-size", fontSize + "px", "important");
  content.style.setProperty("line-height", "1.75", "important");
  content.style.setProperty("color", themes[themeIndex].color, "important");
  content.style.setProperty("background-color", themes[themeIndex].background, "important");
  content.style.setProperty("--sparrow-reader-font-size", fontSize + "px", "important");
  content.style.setProperty("--sparrow-reader-color", themes[themeIndex].color, "important");
  content.style.setProperty("--sparrow-reader-link", themes[themeIndex].link, "important");
  content.style.setProperty("max-width", "760px", "important");
  content.style.setProperty("margin-left", "auto", "important");
  content.style.setProperty("margin-right", "auto", "important");
  content.style.setProperty("padding", "40px 56px", "important");
  content.scrollIntoView({ block: "start" });

  var controls = document.createElement("div");
  controls.id = PANEL_ID;
  controls.setAttribute("role", "toolbar");
  controls.setAttribute("aria-label", "Reading mode controls");
  controls.innerHTML =
    "<span>Reading mode</span>" +
    "<button type='button' data-reader-action='smaller' title='Decrease text size' aria-label='Decrease text size'>A−</button>" +
    "<button type='button' data-reader-action='larger' title='Increase text size' aria-label='Increase text size'>A+</button>" +
    "<button type='button' data-reader-action='theme' title='Change reading theme' aria-label='Change reading theme'>Dark</button>" +
    "<button type='button' data-reader-action='close' title='Exit reading mode' aria-label='Exit reading mode'>×</button>";
  document.body.appendChild(controls);
  controls.addEventListener("click", function (event) {
    var action = event.target && event.target.getAttribute("data-reader-action");
    if (action === "smaller" || action === "larger") {
      fontSize = Math.max(16, Math.min(30, fontSize + (action === "larger" ? 2 : -2)));
      content.style.setProperty("--sparrow-reader-font-size", fontSize + "px", "important");
    } else if (action === "theme") {
      themeIndex = (themeIndex + 1) % themes.length;
      var theme = themes[themeIndex];
      content.style.setProperty("color", theme.color, "important");
      content.style.setProperty("background-color", theme.background, "important");
      content.style.setProperty("--sparrow-reader-color", theme.color, "important");
      content.style.setProperty("--sparrow-reader-link", theme.link, "important");
      event.target.textContent = theme.name;
    } else if (action === "close") {
      document.dispatchEvent(new KeyboardEvent("keydown", {
        key: "r", ctrlKey: true, shiftKey: true, bubbles: true, cancelable: true
      }));
    }
  });
})();`
}
