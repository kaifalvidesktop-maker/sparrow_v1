package main

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type AutocompleteResult struct {
	Text   string `json:"text"`
	URL    string `json:"url"`
	Title  string `json:"title"`
	Source string `json:"source"` // "history" | "bookmark" | "search"
}

var suggestHTTPClient = &http.Client{Timeout: 800 * time.Millisecond}

// fetchSearchEngineSuggestions calls Google's public suggest endpoint,
// the same one Chrome uses for the address-bar dropdown.
func fetchSearchEngineSuggestions(query string) []AutocompleteResult {
	if strings.TrimSpace(query) == "" {
		return nil
	}
	endpoint := "https://suggestqueries.google.com/complete/search?client=firefox&q=" + url.QueryEscape(query)
	resp, err := suggestHTTPClient.Get(endpoint)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var raw []json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil || len(raw) < 2 {
		return nil
	}
	var phrases []string
	if err := json.Unmarshal(raw[1], &phrases); err != nil {
		return nil
	}

	out := make([]AutocompleteResult, 0, len(phrases))
	for _, p := range phrases {
		out = append(out, AutocompleteResult{
			Text:   p,
			URL:    "https://www.google.com/search?q=" + url.QueryEscape(p),
			Title:  p,
			Source: "search",
		})
	}
	return out
}

func getAutocompleteResults(query string) []AutocompleteResult {
	query = strings.TrimSpace(strings.ToLower(query))

	if query == "" {
		return []AutocompleteResult{}
	}

	results := make([]AutocompleteResult, 0, 10)

	// History suggestions
	history := history.GetRecent(50)

	for i := len(history) - 1; i >= 0; i-- {
		item := history[i]

		if strings.Contains(strings.ToLower(item.URL), query) ||
			strings.Contains(strings.ToLower(item.Title), query) {

			results = append(results, AutocompleteResult{
				Text:   item.URL,
				URL:    item.URL,
				Title:  item.Title,
				Source: "history",
			})

			if len(results) >= 10 {
				return results
			}
		}
	}

	// Bookmarks suggestions
	bookmarks := bookmarkManager.GetAll()

	for i := len(bookmarks) - 1; i >= 0; i-- {
		item := bookmarks[i]

		if strings.Contains(strings.ToLower(item.URL), query) ||
			strings.Contains(strings.ToLower(item.Title), query) {

			found := false
			for _, r := range results {
				if r.URL == item.URL {
					found = true
					break
				}
			}

			if found {
				continue
			}

			results = append(results, AutocompleteResult{
				Text:   item.URL,
				URL:    item.URL,
				Title:  item.Title,
				Source: "bookmark",
			})

			if len(results) >= 10 {
				return results
			}
		}
	}

	// Live search-engine suggestions (Chrome-style), same as typing in
	// the Google search box. Filled in last so local matches rank first.
	if len(results) < 10 {
		for _, s := range fetchSearchEngineSuggestions(query) {
			results = append(results, s)
			if len(results) >= 10 {
				break
			}
		}
	}

	return results
}

func autocompleteSearch(query string) string {
	query = strings.TrimSpace(query)

	if query == "" {
		return ""
	}

	results := getAutocompleteResults(query)

	if len(results) == 0 {
		return ""
	}

	return results[0].URL
}

func autocompleteHTML() string {
	return `
<script>
(function () {
	"use strict";

	const input = document.querySelector('input[type="text"]');
	if (!input) {
		return;
	}

	let box = null;

	function createBox() {
		if (box) {
			return box;
		}

		box = document.createElement("div");
		box.style.position = "fixed";
		box.style.zIndex = "2147483647";
		box.style.background = "#ffffff";
		box.style.border = "1px solid #dadce0";
		box.style.borderRadius = "10px";
		box.style.boxShadow = "0 4px 18px rgba(0,0,0,.15)";
		box.style.overflow = "hidden";
		box.style.display = "none";

		document.body.appendChild(box);
		return box;
	}

	function hideBox() {
		if (box) {
			box.style.display = "none";
		}
	}

	function showResults(results) {
		const b = createBox();
		b.textContent = "";

		if (!results || results.length === 0) {
			hideBox();
			return;
		}

		const rect = input.getBoundingClientRect();
		b.style.left = rect.left + "px";
		b.style.top = (rect.bottom + 4) + "px";
		b.style.width = rect.width + "px";

		results.forEach(function (item) {
			const row = document.createElement("div");
			row.style.padding = "10px 14px";
			row.style.cursor = "pointer";
			row.style.borderBottom = "1px solid #f1f3f4";

			const rowTitle = document.createElement("div");
			rowTitle.style.cssText = "font-size:13px;color:#202124";
			rowTitle.textContent = item.Title || item.URL;
			const rowUrl = document.createElement("div");
			rowUrl.style.cssText = "font-size:11px;color:#6b7280;margin-top:3px";
			rowUrl.textContent = item.URL;
			row.appendChild(rowTitle);
			row.appendChild(rowUrl);

			row.onmouseenter = function () {
				row.style.background = "#f1f3f4";
			};

			row.onmouseleave = function () {
				row.style.background = "#ffffff";
			};

			row.onclick = function () {
				input.value = item.URL;
				hideBox();

				if (window.realNavigate) {
					window.realNavigate(item.URL);
				}
			};

			b.appendChild(row);
		});

		b.style.display = "block";
	}

	function escapeHTML(value) {
		return String(value)
			.replace(/&/g, "&amp;")
			.replace(/</g, "&lt;")
			.replace(/>/g, "&gt;")
			.replace(/"/g, "&quot;")
			.replace(/'/g, "&#039;");
	}

	input.addEventListener("input", async function () {
		const q = input.value.trim();
		if (!q) {
			hideBox();
			return;
		}

		try {
			if (window.getAutocomplete) {
				const results = await window.getAutocomplete(q);
				showResults(results);
			}
		} catch (e) {
			hideBox();
		}
	});

	input.addEventListener("keydown", function (event) {
		if (event.key === "Escape") {
			hideBox();
		}
	});

	document.addEventListener("click", function (event) {
		if (event.target !== input && (!box || !box.contains(event.target))) {
			hideBox();
		}
	});

	window.addEventListener("resize", hideBox);
})();
</script>
`
}
