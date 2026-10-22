package main

// ---------------------------------------------------
// PER-SITE PERMISSIONS PAGE (sparrow://site-settings)
// ---------------------------------------------------

const siteSettingsPageHTML = `
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<title>SPARROW Site Settings</title>
<style>
  * { box-sizing: border-box; }
  body {
    margin: 0;
    background: #0b0b0d;
    color: #e8e8e8;
    font-family: 'Segoe UI', sans-serif;
    padding: 40px 60px;
  }
  h1 { font-size: 26px; margin-bottom: 6px; }
  .sub { color: #888; font-size: 13px; margin-bottom: 24px; }
  .search {
    width: 320px;
    padding: 10px 16px;
    border-radius: 20px;
    border: 1px solid #2c2c2e;
    background: #17181a;
    color: #fff;
    outline: none;
    margin-bottom: 24px;
    font-size: 13px;
  }
  .search:focus { border-color: #4d90fe; }
  .site-card {
    background: #17181a;
    border: 1px solid #232323;
    border-radius: 12px;
    padding: 16px 20px;
    margin-bottom: 10px;
    box-shadow: 0 4px 14px rgba(0,0,0,0.3);
  }
  .site-domain { font-size: 14px; color: #fff; margin-bottom: 12px; font-weight: 600; }
  .perm-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 6px 0;
    font-size: 12px;
  }
  .perm-label { color: #aaa; }
  select.perm-select {
    background: #1f2023;
    color: #fff;
    border: 1px solid #2c2c2e;
    border-radius: 6px;
    padding: 5px 10px;
    font-size: 11px;
    outline: none;
  }
  .empty { color: #666; font-size: 13px; margin-top: 40px; text-align: center; }
  .reset-all {
    background: #232323;
    color: #ccc;
    border: 1px solid #2c2c2e;
    border-radius: 8px;
    padding: 8px 16px;
    font-size: 12px;
    cursor: pointer;
    margin-bottom: 20px;
  }
  .reset-all:hover { background: #2a1414; color: #ff6b6b; }
</style>
</head>
<body>
  <h1>Site Settings</h1>
  <div class="sub">Manage camera, microphone, location, and notification permissions per website</div>
  <button class="reset-all" onclick="resetAllSites()">Reset All Site Permissions</button>
  <div id="siteList"></div>

<script>
  async function loadSites() {
    const allPerms = await window.getAllSitePermissions();
    const container = document.getElementById('siteList');
    const domains = Object.keys(allPerms);

    if (domains.length === 0) {
      container.innerHTML = '<div class="empty">No sites have requested permissions yet.</div>';
      return;
    }

    container.innerHTML = domains.map(function(domain) {
      const p = allPerms[domain];
      return '<div class="site-card">' +
        '<div class="site-domain">' + domain + '</div>' +
        permRow(domain, 'camera', 'Camera', p.camera) +
        permRow(domain, 'microphone', 'Microphone', p.microphone) +
        permRow(domain, 'location', 'Location', p.location) +
        permRow(domain, 'notifications', 'Notifications', p.notifications) +
      '</div>';
    }).join('');
  }

  function permRow(domain, type, label, value) {
    return '<div class="perm-row">' +
      '<span class="perm-label">' + label + '</span>' +
      '<select class="perm-select" onchange="updatePerm(\'' + domain + '\',\'' + type + '\',this.value)">' +
        '<option value="ask"' + (value === 'ask' ? ' selected' : '') + '>Ask</option>' +
        '<option value="allow"' + (value === 'allow' ? ' selected' : '') + '>Allow</option>' +
        '<option value="deny"' + (value === 'deny' ? ' selected' : '') + '>Block</option>' +
      '</select>' +
    '</div>';
  }

  async function updatePerm(domain, type, value) {
    await window.setSitePermission(domain, type, value);
  }

  async function resetAllSites() {
    if (!confirm('Reset permissions for all sites?')) return;
    await window.resetAllPermissions();
    loadSites();
  }

  loadSites();
</script>
</body>
</html>
`
