package main

import (
	"os"
	"strings"
	"testing"
)

func TestAppLockPersistenceAndUnlock(t *testing.T) {
	tempDir := t.TempDir()
	oldAppData, hadAppData := os.LookupEnv("APPDATA")
	oldXDG, hadXDG := os.LookupEnv("XDG_CONFIG_HOME")
	if err := os.Setenv("APPDATA", tempDir); err != nil {
		t.Fatal(err)
	}

	if err := os.Setenv("XDG_CONFIG_HOME", tempDir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if hadAppData {
			_ = os.Setenv("APPDATA", oldAppData)
		} else {
			_ = os.Unsetenv("APPDATA")
		}
		if hadXDG {
			_ = os.Setenv("XDG_CONFIG_HOME", oldXDG)
		} else {
			_ = os.Unsetenv("XDG_CONFIG_HOME")
		}
	}()

	previous := appLock.Snapshot()
	defer appLock.Restore(previous)
	appLock.Restore(AppLockSnapshot{})

	const password = "test-password"
	if err := appLock.Configure("   "); err == nil {
		t.Fatal("expected an empty password to be rejected")
	}
	if err := appLock.Configure(password); err != nil {
		t.Fatalf("configure app lock: %v", err)
	}
	if appLock.IsLocked() {
		t.Fatal("newly configured lock should be unlocked until explicitly locked")
	}

	path, err := stateFilePath()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read persisted app lock: %v", err)
	}
	if strings.Contains(string(data), password) {
		t.Fatal("password must not be persisted in plaintext")
	}

	state := loadPersistedState()
	if state == nil || !state.AppLock.Enabled || len(state.AppLock.Salt) == 0 || len(state.AppLock.Hash) == 0 {
		t.Fatal("expected salted app-lock verifier to be persisted")
	}
	appLock.Restore(state.AppLock)
	if !appLock.IsLocked() {
		t.Fatal("restoring an enabled app lock must require unlocking")
	}
	if appLock.Unlock("wrong-password") {
		t.Fatal("incorrect password unexpectedly unlocked app")
	}
	if !appLock.Unlock(password) {
		t.Fatal("correct password did not unlock app")
	}

	if err := appLock.Disable("wrong-password"); err == nil {
		t.Fatal("expected an incorrect password to be rejected when disabling lock")
	}
	if err := appLock.Disable(password); err != nil {
		t.Fatalf("disable app lock: %v", err)
	}
	if appLock.IsEnabled() || appLock.IsLocked() {
		t.Fatal("app lock should be disabled after successful password confirmation")
	}
}

func TestAppLockInternalPageRoute(t *testing.T) {
	if page, ok := internalPageHTML("sparrow://lock"); !ok || page != appLockPageHTML {
		t.Fatal("expected sparrow://lock to resolve to the app-lock page")
	}
}
