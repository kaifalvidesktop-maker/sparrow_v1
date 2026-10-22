package main

import (
	"strings"
	"sync"
	"time"
)

// ---------------------------------------------------
// COOKIE MANAGER (per-site cookie tracking + control)
// Note: real cookie interception requires deep engine hooks;
// this tracks cookies SPARROW is aware should be scoped per-site
// and lets the user block/clear them per domain.
// ---------------------------------------------------

type CookieRecord struct {
	Domain string
	Name   string
	SetAt  time.Time
}

type CookieReport struct {
	Domain string
	Name   string
}

type CookieManager struct {
	mu      sync.Mutex
	Records map[string][]CookieRecord // domain -> cookies
	Blocked map[string]bool           // domain -> cookies blocked entirely
}

var cookieManager = &CookieManager{
	Records: make(map[string][]CookieRecord),
	Blocked: make(map[string]bool),
}

// Record that a cookie was set for a domain (called from JS hook)
func (cm *CookieManager) Record(domain string, name string) bool {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	domain = strings.ToLower(strings.TrimSpace(domain))
	name = strings.TrimSpace(name)
	if name == "" {
		return false
	}
	if cm.Blocked[domain] {
		return false // rejected
	}
	records := cm.Records[domain]
	for i := range records {
		if records[i].Name == name {
			records[i].SetAt = time.Now()
			return true
		}
	}
	cm.Records[domain] = append(records, CookieRecord{
		Domain: domain,
		Name:   name,
		SetAt:  time.Now(),
	})
	if len(cm.Records[domain]) > 100 {
		cm.Records[domain] = cm.Records[domain][len(cm.Records[domain])-100:]
	}
	return true
}

// ClearDomain removes all tracked cookies for one domain
func (cm *CookieManager) ClearDomain(domain string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.Records, strings.ToLower(domain))
}

// ClearAll wipes every tracked cookie
func (cm *CookieManager) ClearAll() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.Records = make(map[string][]CookieRecord)
}

// SetBlocked toggles whether a domain is allowed to set cookies
func (cm *CookieManager) SetBlocked(domain string, blocked bool) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.Blocked[strings.ToLower(domain)] = blocked
}

// IsBlocked checks if a domain's cookies are blocked
func (cm *CookieManager) IsBlocked(domain string) bool {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	return cm.Blocked[strings.ToLower(domain)]
}

// GetSummary returns domain -> cookie count (for the cookies UI page)
func (cm *CookieManager) GetSummary() map[string]int {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	summary := make(map[string]int)
	for domain, records := range cm.Records {
		seen := map[string]bool{}
		for _, record := range records {
			if record.Name == "" {
				continue
			}
			seen[record.Name] = true
		}
		summary[domain] = len(seen)
	}
	return summary
}

// The cookie hook batches unique cookie names to reduce blocking Go calls.
const cookieHookJS = `
(function() {
	if (window.__sparrowCookieHooked) return;
	window.__sparrowCookieHooked = true;
	try {
		var originalDescriptor = Object.getOwnPropertyDescriptor(Document.prototype, 'cookie');
		if (!originalDescriptor || !originalDescriptor.get || !originalDescriptor.set) return;
		var seen = Object.create(null);
		var seenCount = 0;
		var pending = [];
		var timer = null;

		function flush() {
			timer = null;
			var batch = pending;
			pending = [];
			if (!batch.length || !window.reportCookies) return;
			try { window.reportCookies(batch); } catch (e) {}
		}

		function queue(domain, value) {
			var name = String(value).split('=')[0].trim();
			if (!name) return;
			var key = domain + '|' + name;
			if (seen[key]) return;
			if (seenCount > 500) {
				seen = Object.create(null);
				seenCount = 0;
			}
			seen[key] = true;
			seenCount++;
			pending.push({ Domain: domain, Name: name });
			if (pending.length >= 50) {
				if (timer) clearTimeout(timer);
				flush();
			} else if (!timer) {
				timer = setTimeout(flush, 1500);
			}
		}

		Object.defineProperty(document, 'cookie', {
			configurable: true,
			enumerable: true,
			get: function() { return originalDescriptor.get.call(document); },
			set: function(val) {
				try { queue(window.location.hostname, val); } catch (e) {}
				originalDescriptor.set.call(document, val);
			}
		});
		window.addEventListener('pagehide', flush, { once: true });
	} catch (e) {}
})();
`

const cookiesPageHTML = `
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<title>SPARROW Cookies</title>
<style>
  * { box-sizing: border-box; }
  body {
    margin: 0; background: #0b0b0d; color: #e8e8e8;
    font-family: 'Segoe UI', sans-serif; padding: 40px 60px;
  }
  h1 { font-size: 26px; margin-bottom: 6px; }
  .sub { color: #888; font-size: 13px; margin-bottom: 24px; }
  .btn {
    background: #232323; color: #ccc; border: 1px solid #2c2c2e;
    padding: 8px 16px; border-radius: 8px; cursor: pointer; font-size: 12px; margin-bottom: 20px;
  }
  .btn:hover { background: #2a1414; color: #ff6b6b; }
  .cookie-row {
    display: flex; justify-content: space-between; align-items: center;
    background: #17181a; border: 1px solid #232323; border-radius: 10px;
    padding: 12px 16px; margin-bottom: 8px;
  }
  .domain { font-size: 13px; color: #fff; }
  .count { font-size: 11px; color: #888; }
  .row-actions { display: flex; gap: 8px; }
  .mini-btn {
    background: #232323; color: #999; border: 1px solid #2c2c2e;
    padding: 5px 10px; border-radius: 6px; font-size: 11px; cursor: pointer;
  }
  .mini-btn:hover { color: #fff; }
  .empty { color: #666; font-size: 13px; margin-top: 40px; text-align: center; }
</style>
</head>
<body>
  <h1>Cookies</h1>
  <div class="sub">Cookies SPARROW has observed sites setting during this session</div>
  <button class="btn" onclick="clearAllCookies()">Clear All Cookies</button>
  <div id="cookieList"></div>

<script>
  async function loadCookies() {
    const summary = await window.getCookieSummary();
    const domains = Object.keys(summary);
    const container = document.getElementById('cookieList');

    if (domains.length === 0) {
      container.innerHTML = '<div class="empty">No cookies observed yet.</div>';
      return;
    }

    container.innerHTML = domains.map(function(d) {
      return '<div class="cookie-row">' +
        '<div><div class="domain">' + d + '</div><div class="count">' + summary[d] + ' cookie(s)</div></div>' +
        '<div class="row-actions">' +
          '<div class="mini-btn" onclick="clearDomain(\'' + d + '\')">Clear</div>' +
          '<div class="mini-btn" onclick="blockDomain(\'' + d + '\')">Block Future</div>' +
        '</div>' +
      '</div>';
    }).join('');
  }

  async function clearDomain(d) {
    await window.clearDomainCookies(d);
    loadCookies();
  }

  async function blockDomain(d) {
    await window.setCookieBlocked(d, true);
    alert(d + ' will no longer be allowed to set cookies');
  }

  async function clearAllCookies() {
    if (!confirm('Clear all tracked cookies?')) return;
    await window.clearAllCookies();
    loadCookies();
  }

  loadCookies();
</script>
</body>
</html>
`
