package main

import (
	"crypto/rand"
	"crypto/subtle"
	"errors"
	"strings"
	"sync"
)

type AppLockSnapshot struct {
	Enabled bool   `json:"enabled"`
	Salt    []byte `json:"salt,omitempty"`
	Hash    []byte `json:"hash,omitempty"`
}

type AppLock struct {
	mu       sync.Mutex
	enabled  bool
	unlocked bool
	salt     []byte
	hash     []byte
}

var appLock = &AppLock{unlocked: true}

const appLockPageHTML = `
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>SPARROW App Lock</title>
<style>
  * { box-sizing: border-box; }
  body { margin: 0; min-height: 100vh; display: grid; place-items: center; background: #0b0b0d; color: #eef2ff; font-family: 'Segoe UI', sans-serif; }
  main { width: min(440px, calc(100% - 32px)); padding: 32px; background: #17181a; border: 1px solid #2c2d30; border-radius: 18px; box-shadow: 0 20px 50px rgba(0,0,0,.35); }
  .brand { color: #93c5fd; font-size: 12px; font-weight: 700; letter-spacing: .18em; }
  h1 { margin: 10px 0 8px; font-size: 25px; }
  p { color: #a9adb7; font-size: 14px; line-height: 1.5; }
  label { display: block; margin: 18px 0 7px; color: #c9cbd1; font-size: 13px; }
  input { width: 100%; padding: 12px; border: 1px solid #35373d; border-radius: 9px; background: #101113; color: #fff; font-size: 14px; outline: none; }
  input:focus { border-color: #4d90fe; }
  button { width: 100%; margin-top: 14px; padding: 11px 14px; border: 0; border-radius: 9px; background: #3978e8; color: #fff; font-size: 14px; font-weight: 600; cursor: pointer; }
  button.secondary { background: #292b30; }
  button.danger { background: #6f3437; }
  button:hover { filter: brightness(1.1); }
  [hidden] { display: none !important; }
  #message { min-height: 20px; margin-top: 12px; color: #fca5a5; font-size: 13px; }
</style>
</head>
<body>
<main>
  <div class="brand">SPARROW</div>
  <h1>App lock</h1>
  <p id="intro">Set a password to protect SPARROW each time it starts.</p>
  <section id="setup" hidden>
    <label for="new-password">Password</label>
    <input id="new-password" type="password" autocomplete="new-password">
    <label for="confirm-password">Confirm password</label>
    <input id="confirm-password" type="password" autocomplete="new-password">
    <button id="save-password">Save password</button>
  </section>
  <section id="unlock" hidden>
    <label for="current-password">Password</label>
    <input id="current-password" type="password" autocomplete="current-password">
    <button id="unlock-app">Unlock app</button>
  </section>
  <section id="manage" hidden>
    <p>Your password is saved. SPARROW will ask for it the next time the app starts.</p>
    <button class="secondary" id="lock-now">Lock now</button>
    <label for="disable-password">To stop requiring a password, confirm it below.</label>
    <input id="disable-password" type="password" autocomplete="current-password">
    <button class="danger" id="disable-lock">Unlock and turn off password</button>
  </section>
  <div id="message" role="status" aria-live="polite"></div>
</main>
<script>
  (async function () {
    var status = await window.getAppLockStatus();
    var locked = status.configured && status.locked;
    document.getElementById('setup').hidden = status.configured;
    document.getElementById('unlock').hidden = !locked;
    document.getElementById('manage').hidden = !status.configured || locked;
    document.getElementById('intro').textContent = locked
      ? 'Enter your password to continue using SPARROW.'
      : status.configured
        ? 'Your app password is active.'
        : 'Set a password to protect SPARROW each time it starts.';

    function showError(error) {
      document.getElementById('message').textContent = error && error.message ? error.message : String(error);
    }
    function clearMessage() {
      document.getElementById('message').textContent = '';
    }

    document.getElementById('save-password').onclick = async function () {
      clearMessage();
      var password = document.getElementById('new-password').value;
      if (!password.trim()) {
        showError('Enter a password.');
        return;
      }
      if (password !== document.getElementById('confirm-password').value) {
        showError('The passwords do not match.');
        return;
      }
      try {
        await window.setAppLockPassword(password);
        await window.lockApp();
      } catch (error) {
        showError(error);
      }
    };

    document.getElementById('unlock-app').onclick = async function () {
      clearMessage();
      try {
        var unlocked = await window.unlockApp(document.getElementById('current-password').value);
        if (!unlocked) showError('Incorrect password.');
      } catch (error) {
        showError(error);
      }
    };

    document.getElementById('lock-now').onclick = async function () {
      clearMessage();
      try {
        await window.lockApp();
      } catch (error) {
        showError(error);
      }
    };

    document.getElementById('disable-lock').onclick = async function () {
      clearMessage();
      try {
        await window.disableAppLock(document.getElementById('disable-password').value);
      } catch (error) {
        showError(error);
      }
    };

    var password = document.getElementById(locked ? 'current-password' : 'new-password');
    if (password) password.focus();
  })().catch(function (error) {
    document.getElementById('message').textContent = error && error.message ? error.message : String(error);
  });
</script>
</body>
</html>
`

func (l *AppLock) Snapshot() AppLockSnapshot {
	l.mu.Lock()
	defer l.mu.Unlock()
	return AppLockSnapshot{
		Enabled: l.enabled,
		Salt:    append([]byte(nil), l.salt...),
		Hash:    append([]byte(nil), l.hash...),
	}
}

func (l *AppLock) Restore(snapshot AppLockSnapshot) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.enabled = snapshot.Enabled && len(snapshot.Salt) > 0 && len(snapshot.Hash) > 0
	l.unlocked = !l.enabled
	l.salt = append([]byte(nil), snapshot.Salt...)
	l.hash = append([]byte(nil), snapshot.Hash...)
	if !l.enabled {
		l.salt = nil
		l.hash = nil
	}
}

func (l *AppLock) Status() map[string]bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return map[string]bool{
		"configured": l.enabled,
		"locked":     l.enabled && !l.unlocked,
	}
}

func (l *AppLock) IsEnabled() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.enabled
}

func (l *AppLock) IsLocked() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.enabled && !l.unlocked
}

func (l *AppLock) Configure(password string) error {
	if strings.TrimSpace(password) == "" {
		return errors.New("password cannot be empty")
	}

	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return err
	}
	hash := deriveKeyWithSalt(password, salt)

	l.mu.Lock()
	previous := AppLockSnapshot{
		Enabled: l.enabled,
		Salt:    append([]byte(nil), l.salt...),
		Hash:    append([]byte(nil), l.hash...),
	}
	l.enabled = true
	l.unlocked = true
	l.salt = salt
	l.hash = hash
	l.mu.Unlock()

	if err := savePersistedState(); err != nil {
		l.Restore(previous)
		return err
	}
	return nil
}

func (l *AppLock) Unlock(password string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if !l.enabled || len(l.salt) == 0 || len(l.hash) == 0 {
		return false
	}
	candidate := deriveKeyWithSalt(password, l.salt)
	if subtle.ConstantTimeCompare(candidate, l.hash) != 1 {
		return false
	}
	l.unlocked = true
	return true
}

func (l *AppLock) Lock() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if !l.enabled {
		return false
	}
	l.unlocked = false
	return true
}

func (l *AppLock) Disable(password string) error {
	l.mu.Lock()
	if !l.enabled {
		l.mu.Unlock()
		return nil
	}
	candidate := deriveKeyWithSalt(password, l.salt)
	if subtle.ConstantTimeCompare(candidate, l.hash) != 1 {
		l.mu.Unlock()
		return errors.New("incorrect password")
	}
	previous := AppLockSnapshot{
		Enabled: l.enabled,
		Salt:    append([]byte(nil), l.salt...),
		Hash:    append([]byte(nil), l.hash...),
	}
	l.enabled = false
	l.unlocked = true
	l.salt = nil
	l.hash = nil
	l.mu.Unlock()

	if err := savePersistedState(); err != nil {
		l.Restore(previous)
		return err
	}
	return nil
}
