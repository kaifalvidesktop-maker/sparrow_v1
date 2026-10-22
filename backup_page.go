package main

const backupPageHTML = `
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<title>SPARROW Backup & Restore</title>
<style>
  * { box-sizing: border-box; }
  body {
    margin: 0; background: #0b0b0d; color: #e8e8e8;
    font-family: 'Segoe UI', sans-serif; padding: 40px 60px;
  }
  h1 { font-size: 26px; margin-bottom: 6px; }
  .sub { color: #888; font-size: 13px; margin-bottom: 24px; max-width: 480px; }
  .card {
    background: #17181a; border: 1px solid #232323; border-radius: 14px;
    padding: 24px; max-width: 480px; margin-bottom: 20px; box-shadow: 0 6px 20px rgba(0,0,0,0.35);
  }
  .card-title { font-size: 14px; font-weight: 600; color: #fff; margin-bottom: 6px; }
  .card-desc { font-size: 12px; color: #888; margin-bottom: 16px; }
  .btn {
    background: #4d90fe; color: #000; border: none; padding: 10px 20px;
    border-radius: 8px; font-size: 13px; font-weight: 600; cursor: pointer;
  }
  .btn-secondary {
    background: #232323; color: #ccc; border: 1px solid #2c2c2e;
    padding: 10px 20px; border-radius: 8px; font-size: 13px; cursor: pointer;
  }
  .status { font-size: 12px; color: #3ddc84; margin-top: 12px; display: none; }
  .status.show { display: block; }
</style>
</head>
<body>
  <h1>Backup & Restore</h1>
  <div class="sub">Move your bookmarks, favorites, history, downloads, settings, and saved information between computers.</div>

  <div class="card">
    <div class="card-title">Create Backup</div>
    <div class="card-desc">Saves your bookmarks, favorites, history, download records, shortcuts, settings, autofill, and appearance preferences to a JSON file in your Downloads folder. Passwords are not included; use the separate encrypted vault export.</div>
    <button class="btn" onclick="createBackup()">Save Backup File</button>
    <div class="status" id="backupStatus"></div>
  </div>

  <div class="card">
    <div class="card-title">Restore from Backup</div>
    <div class="card-desc">Select a SPARROW backup file to add its saved data and apply its settings. Existing data is kept, and duplicates are skipped.</div>
    <button class="btn-secondary" onclick="document.getElementById('restoreFile').click()">Choose Backup File</button>
    <input type="file" id="restoreFile" accept=".json" style="display:none;" onchange="restoreBackupFileSelected()">
    <div class="status" id="restoreStatus"></div>
  </div>

<script>
  function showStatus(statusEl, message) {
    statusEl.innerText = message;
    statusEl.classList.add('show');
  }

  async function createBackup() {
    const statusEl = document.getElementById('backupStatus');
    showStatus(statusEl, 'Creating backup...');
    try {
      const path = await window.saveBackupToFile();
      showStatus(statusEl, 'Saved to: ' + path);
    } catch (error) {
      console.error('SPARROW could not save the backup:', error);
      showStatus(statusEl, 'Could not save backup: ' + (error && error.message ? error.message : String(error)));
    }
  }

  function restoreBackupFileSelected() {
    const input = document.getElementById('restoreFile');
    if (!input || !input.files || input.files.length === 0) {
      return;
    }
    const file = input.files[0];
    if (!file) return;
    const reader = new FileReader();
    const statusEl = document.getElementById('restoreStatus');
    if (!statusEl) {
      console.error('SPARROW restore status element is missing.');
      return;
    }
    showStatus(statusEl, 'Reading backup file...');
    reader.onload = async function(e) {
      await restoreBackupContents(e.target.result, statusEl);
    };
    reader.onerror = function() {
      showStatus(statusEl, 'Could not read the selected backup file.');
    };
    reader.onloadend = function() {
      input.value = '';
    };
    reader.readAsText(file);
  }

  async function restoreBackupContents(contents, statusEl) {
    showStatus(statusEl, 'Restoring backup...');
    try {
      const count = await window.restoreBackup(contents);
      if (window.__sparrowRefreshChrome) await window.__sparrowRefreshChrome();
      showStatus(statusEl, count + ' data item(s) restored. Settings and appearance preferences applied.');
    } catch (error) {
      console.error('SPARROW could not restore the backup:', error);
      showStatus(statusEl, 'Could not restore backup: ' + (error && error.message ? error.message : String(error)));
    }
  }
</script>
</body>
</html>
`
