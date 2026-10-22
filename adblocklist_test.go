package main

import "testing"

func TestAdBlocklistHasExactly200UniqueDomains(t *testing.T) {
	if len(adBlocklist) != 200 {
		t.Fatalf("adBlocklist length = %d, want 200", len(adBlocklist))
	}

	seen := map[string]bool{}
	for _, domain := range adBlocklist {
		if domain == "" {
			t.Fatalf("adBlocklist contains empty domain")
		}
		if seen[domain] {
			t.Fatalf("duplicate domain in adBlocklist: %s", domain)
		}
		seen[domain] = true
	}

	if len(seen) != 200 {
		t.Fatalf("adBlocklist unique domains = %d, want 200", len(seen))
	}
}

func TestIsBlockedMatchesHostnamesWithoutSubstringMatches(t *testing.T) {
	if !IsBlocked("ads.example.com") {
		t.Fatal("expected ads.example.com to be blocked when example.com is in the list")
	}
	if IsBlocked("notexample.com") {
		t.Fatal("notexample.com should not match example.com")
	}
	if IsBlocked("localhost") {
		t.Fatal("localhost must never be blocked")
	}
	if IsBlocked("127.0.0.1") {
		t.Fatal("127.0.0.1 must never be blocked")
	}
	if IsBlocked("::1") {
		t.Fatal("::1 must never be blocked")
	}
}
