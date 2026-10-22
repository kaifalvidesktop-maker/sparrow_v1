package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"strings"
	"sync"
	"time"
)

// ---------------------------------------------------
// ENCRYPTED PASSWORD MANAGER
// Master password derives an AES-256 key (via SHA-256).
// Passwords remain encrypted while stored on disk and are only decrypted
// after the user unlocks the vault.
// ---------------------------------------------------

type PasswordEntry struct {
	ID        int
	Domain    string
	Username  string
	encrypted []byte // AES-GCM ciphertext, never exposed directly
	nonce     []byte
	CreatedAt time.Time
}

type PasswordVault struct {
	mu          sync.Mutex
	Entries     []*PasswordEntry
	NextID      int
	masterKey   []byte // derived key, held only while unlocked
	salt        []byte // random per-vault salt persisted with encrypted entries
	Unlocked    bool
	FailedTries int
}

var passwordVault = &PasswordVault{
	Entries: []*PasswordEntry{},
	NextID:  1,
}

// deriveKey is kept for legacy vaults created before salted derivation.
func deriveKey(masterPassword string) []byte {
	sum := sha256.Sum256([]byte(masterPassword))
	return sum[:]
}

func deriveKeyWithSalt(masterPassword string, salt []byte) []byte {
	data := append(append([]byte{}, salt...), []byte(masterPassword)...)
	sum := sha256.Sum256(data)
	key := sum[:]
	// Iteration hardens the key against offline guessing without adding a dependency.
	for i := 0; i < 200000; i++ {
		n := sha256.Sum256(key)
		key = n[:]
	}
	return append([]byte(nil), key...)
}

// Unlock sets the vault's active key from the master password.
// When saved entries exist, the first entry verifies the password.
func (pv *PasswordVault) Unlock(masterPassword string) bool {
	pv.mu.Lock()
	defer pv.mu.Unlock()

	if strings.TrimSpace(masterPassword) == "" {
		return false
	}

	if len(pv.salt) == 0 {
		if len(pv.Entries) > 0 {
			// Legacy vault: retain compatibility with the original key format.
			pv.masterKey = deriveKey(masterPassword)
		} else {
			pv.salt = make([]byte, 16)
			if _, err := rand.Read(pv.salt); err != nil {
				return false
			}
			pv.masterKey = deriveKeyWithSalt(masterPassword, pv.salt)
		}
	} else {
		pv.masterKey = deriveKeyWithSalt(masterPassword, pv.salt)
	}
	key := pv.masterKey
	if len(pv.Entries) > 0 {
		if _, err := decryptWithKey(key, pv.Entries[0].encrypted, pv.Entries[0].nonce); err != nil {
			pv.FailedTries++
			return false
		}
	}
	pv.masterKey = key
	pv.Unlocked = true
	pv.FailedTries = 0
	return true
}

// Lock clears the active key from memory (entries remain but unreadable
// until unlocked again with the correct password)
func (pv *PasswordVault) Lock() {
	pv.mu.Lock()
	defer pv.mu.Unlock()
	pv.masterKey = nil
	pv.Unlocked = false
}

// IsUnlocked reports current lock state
func (pv *PasswordVault) IsUnlocked() bool {
	pv.mu.Lock()
	defer pv.mu.Unlock()
	return pv.Unlocked
}

// encrypt uses AES-GCM to encrypt plaintext with the current master key
func (pv *PasswordVault) encrypt(plaintext string) ([]byte, []byte, error) {
	block, err := aes.NewCipher(pv.masterKey)
	if err != nil {
		return nil, nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, nil, err
	}

	ciphertext := gcm.Seal(nil, nonce, []byte(plaintext), nil)
	return ciphertext, nonce, nil
}

// decrypt reverses encrypt() using the current master key
func (pv *PasswordVault) decrypt(ciphertext []byte, nonce []byte) (string, error) {
	return decryptWithKey(pv.masterKey, ciphertext, nonce)
}

func decryptWithKey(key, ciphertext, nonce []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// Save adds a new encrypted password entry. Vault must be unlocked.
func (pv *PasswordVault) Save(domain string, username string, password string) (*PasswordEntry, error) {
	pv.mu.Lock()
	defer pv.mu.Unlock()

	if !pv.Unlocked {
		return nil, errors.New("vault is locked")
	}

	ciphertext, nonce, err := pv.encrypt(password)
	if err != nil {
		return nil, err
	}

	entry := &PasswordEntry{
		ID:        pv.NextID,
		Domain:    strings.ToLower(strings.TrimSpace(domain)),
		Username:  username,
		encrypted: ciphertext,
		nonce:     nonce,
		CreatedAt: time.Now(),
	}

	pv.Entries = append(pv.Entries, entry)
	pv.NextID++
	return entry, nil
}

// Reveal decrypts and returns the plaintext password for one entry.
// This is the ONLY function that ever exposes a decrypted password.
func (pv *PasswordVault) Reveal(id int) (string, error) {
	pv.mu.Lock()
	defer pv.mu.Unlock()

	if !pv.Unlocked {
		return "", errors.New("vault is locked")
	}

	for _, e := range pv.Entries {
		if e.ID == id {
			return pv.decrypt(e.encrypted, e.nonce)
		}
	}
	return "", errors.New("entry not found")
}

// Remove deletes a password entry permanently
func (pv *PasswordVault) Remove(id int) bool {
	pv.mu.Lock()
	defer pv.mu.Unlock()

	newList := []*PasswordEntry{}
	removed := false
	for _, e := range pv.Entries {
		if e.ID == id {
			removed = true
			continue
		}
		newList = append(newList, e)
	}
	pv.Entries = newList
	return removed
}

type StoredPassword struct {
	ID        int
	Domain    string
	Username  string
	Encrypted []byte
	Nonce     []byte
	CreatedAt time.Time
}

func (pv *PasswordVault) Snapshot() []StoredPassword {
	pv.mu.Lock()
	defer pv.mu.Unlock()
	result := make([]StoredPassword, 0, len(pv.Entries))
	for _, entry := range pv.Entries {
		result = append(result, StoredPassword{
			ID: entry.ID, Domain: entry.Domain, Username: entry.Username,
			Encrypted: append([]byte(nil), entry.encrypted...),
			Nonce:     append([]byte(nil), entry.nonce...), CreatedAt: entry.CreatedAt,
		})
	}
	return result
}

func (pv *PasswordVault) Restore(entries []StoredPassword) {
	pv.mu.Lock()
	defer pv.mu.Unlock()
	pv.Entries = make([]*PasswordEntry, 0, len(entries))
	pv.NextID = 1
	for _, entry := range entries {
		pv.Entries = append(pv.Entries, &PasswordEntry{
			ID: entry.ID, Domain: entry.Domain, Username: entry.Username,
			encrypted: append([]byte(nil), entry.Encrypted...),
			nonce:     append([]byte(nil), entry.Nonce...), CreatedAt: entry.CreatedAt,
		})
		if entry.ID >= pv.NextID {
			pv.NextID = entry.ID + 1
		}
	}
}

func (pv *PasswordVault) SaltSnapshot() []byte {
	pv.mu.Lock()
	defer pv.mu.Unlock()
	return append([]byte(nil), pv.salt...)
}

func (pv *PasswordVault) SetSalt(salt []byte) {
	pv.mu.Lock()
	defer pv.mu.Unlock()
	pv.salt = append([]byte(nil), salt...)
}

// GetMetadataOnly returns entries WITHOUT decrypted passwords —
// safe to send to the UI for listing (domain + username only)
type PasswordMetadata struct {
	ID       int
	Domain   string
	Username string
}

func (pv *PasswordVault) GetMetadataOnly() []PasswordMetadata {
	pv.mu.Lock()
	defer pv.mu.Unlock()

	result := []PasswordMetadata{}
	for _, e := range pv.Entries {
		result = append(result, PasswordMetadata{
			ID:       e.ID,
			Domain:   e.Domain,
			Username: e.Username,
		})
	}
	return result
}

// FindForDomain returns metadata for entries matching a domain (for autofill prompt)
func (pv *PasswordVault) FindForDomain(domain string) []PasswordMetadata {
	pv.mu.Lock()
	defer pv.mu.Unlock()

	domain = strings.ToLower(domain)
	result := []PasswordMetadata{}
	for _, e := range pv.Entries {
		if e.Domain == domain {
			result = append(result, PasswordMetadata{ID: e.ID, Domain: e.Domain, Username: e.Username})
		}
	}
	return result
}

// WipeAll destroys the entire vault (used by "Clear All Data" in settings)
func (pv *PasswordVault) WipeAll() {
	pv.mu.Lock()
	defer pv.mu.Unlock()
	pv.Entries = []*PasswordEntry{}
	pv.masterKey = nil
	pv.salt = nil
	pv.Unlocked = false
}

// GeneratePassword creates a strong random password
func GenerateStrongPassword(length int) string {
	if length < 8 {
		length = 16
	}
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	result := make([]byte, length)
	randomBytes := make([]byte, length)
	rand.Read(randomBytes)
	for i, b := range randomBytes {
		result[i] = charset[int(b)%len(charset)]
	}
	return string(result)
}
