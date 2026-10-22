package main

import "html"

// ---------------------------------------------------
// CUSTOM SPARROW ERROR PAGES
// ---------------------------------------------------

func buildBlockedPageHTML(domain string) string {
	return `
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<style>
  body {
    margin: 0; height: 100vh; background: #0b0b0d; color: #e8e8e8;
    font-family: 'Segoe UI', sans-serif; display: flex; flex-direction: column;
    align-items: center; justify-content: center; text-align: center;
  }
  .icon { font-size: 60px; margin-bottom: 20px; }
  h1 { font-size: 22px; margin-bottom: 10px; }
  p { color: #888; font-size: 13px; max-width: 400px; }
  .domain { color: #ff8a8a; font-family: monospace; }
</style>
</head>
<body>
  <div class="icon">&#128737;</div>
  <h1>Blocked by SPARROW Security Shield</h1>
  <p>The domain <span class="domain">` + html.EscapeString(domain) + `</span> was blocked because it matches a known ad or tracker pattern. You can adjust this in Settings if it was blocked by mistake.</p>
</body>
</html>
`
}

func buildOfflinePageHTML(target string) string {
	return buildErrorPageHTML("offline", target, "")
}

const notFoundPageHTML = `
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<style>
  body {
    margin: 0; height: 100vh; background: #0b0b0d; color: #e8e8e8;
    font-family: 'Segoe UI', sans-serif; display: flex; flex-direction: column;
    align-items: center; justify-content: center; text-align: center;
  }
  .icon { font-size: 60px; margin-bottom: 20px; opacity: 0.5; }
  h1 { font-size: 22px; margin-bottom: 10px; }
  p { color: #888; font-size: 13px; }
  button {
    margin-top: 20px; background: #232323; color: #ccc; border: 1px solid #2c2c2e;
    padding: 10px 24px; border-radius: 8px; cursor: pointer; font-size: 13px;
  }
  button:hover { background: #2a2a2c; color: #fff; }
</style>
</head>
<body>
  <div class="icon">&#128269;</div>
  <h1>Page Not Found</h1>
  <p>SPARROW couldn't find that page. It may have been moved or no longer exists.</p>
  <button onclick="goHome()">Go Home</button>
</body>
</html>
`
