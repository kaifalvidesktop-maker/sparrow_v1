package main

import "testing"

func TestFileURLFromSlashPath(t *testing.T) {
	got := fileURLFromSlashPath("C:/Users/me/My File.pdf")
	if got != "file:///C:/Users/me/My%20File.pdf" {
		t.Fatalf("got %s", got)
	}
}

func TestLocalPDFPathHandlesDriveLetterAndBrokenHost(t *testing.T) {
	cases := map[string]string{
		"file:///C:/Users/me/My%20File.pdf": "C:/Users/me/My File.pdf",
		"file://C:/Users/me/a.pdf":          "C:/Users/me/a.pdf", // old broken form
	}
	for in, want := range cases {
		p, ok := sparrowLocalPDFPath(in)
		if !ok || p != want {
			t.Errorf("%s: got %q ok=%v want %q", in, p, ok, want)
		}
	}
}
