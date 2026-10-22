package main

import (
	"sync"
	"time"
)

// ---------------------------------------------------
// FEEDBACK / BUG REPORT (stored in RAM, for Kaif to review
// and export during development — no network calls made)
// ---------------------------------------------------

type FeedbackEntry struct {
	ID        int
	Message   string
	Category  string // "bug", "idea", "praise"
	CreatedAt time.Time
}

type FeedbackManager struct {
	mu     sync.Mutex
	Items  []*FeedbackEntry
	NextID int
}

var feedbackManager = &FeedbackManager{
	Items:  []*FeedbackEntry{},
	NextID: 1,
}

// Submit adds a new feedback entry
func (fm *FeedbackManager) Submit(message string, category string) *FeedbackEntry {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	valid := map[string]bool{"bug": true, "idea": true, "praise": true}
	if !valid[category] {
		category = "idea"
	}

	entry := &FeedbackEntry{
		ID:        fm.NextID,
		Message:   message,
		Category:  category,
		CreatedAt: time.Now(),
	}
	fm.Items = append(fm.Items, entry)
	fm.NextID++
	return entry
}

// GetAll returns all feedback entries collected this session
func (fm *FeedbackManager) GetAll() []*FeedbackEntry {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	return fm.Items
}

const feedbackPageHTML = `
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<title>SPARROW Feedback</title>
<style>
  * { box-sizing: border-box; }
  body {
    margin: 0; background: #0b0b0d; color: #e8e8e8;
    font-family: 'Segoe UI', sans-serif; padding: 40px 60px;
  }
  h1 { font-size: 26px; margin-bottom: 6px; }
  .sub { color: #888; font-size: 13px; margin-bottom: 24px; max-width: 460px; }
  .card {
    background: #17181a; border: 1px solid #232323; border-radius: 14px;
    padding: 24px; max-width: 460px; box-shadow: 0 6px 20px rgba(0,0,0,0.35);
  }
  select, textarea {
    width: 100%; background: #1f2023; border: 1px solid #2c2c2e; border-radius: 8px;
    padding: 10px 12px; color: #fff; font-size: 13px; outline: none; font-family: inherit;
  }
  textarea { min-height: 100px; resize: vertical; margin-top: 12px; }
  .btn {
    background: #4d90fe; color: #000; border: none; padding: 10px 22px;
    border-radius: 8px; font-size: 13px; font-weight: 600; cursor: pointer; margin-top: 14px;
  }
</style>
</head>
<body>
  <h1>Send Feedback</h1>
  <div class="sub">Found a bug or have an idea for SPARROW? This stays on your device this session — nothing is sent anywhere.</div>

  <div class="card">
    <select id="fb-category">
      <option value="bug">Bug Report</option>
      <option value="idea">Feature Idea</option>
      <option value="praise">Something I Like</option>
    </select>
    <textarea id="fb-message" placeholder="Describe what happened or what you'd like to see..."></textarea>
    <button class="btn" onclick="submitFeedback()">Submit</button>
  </div>

<script>
  async function submitFeedback() {
    const category = document.getElementById('fb-category').value;
    const message = document.getElementById('fb-message').value;
    if (message.trim() === '') return;
    await window.submitFeedback(message, category);
    document.getElementById('fb-message').value = '';
    alert('Thanks! Saved for this session.');
  }
</script>
</body>
</html>
`
