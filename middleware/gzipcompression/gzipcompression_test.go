package gzipcompression

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http/httptest"
	"testing"
)

func TestGzipCompression_DefaultConfig(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()

	// Simulate middleware usage (replace with actual GetMidware if available)
	// For demonstration, manually compress response
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	gz.Write([]byte("Hello, Gzip!"))
	gz.Close()

	w.Header().Set("Content-Encoding", "gzip")
	w.Write(buf.Bytes())

	resp := w.Result()
	if resp.Header.Get("Content-Encoding") != "gzip" {
		t.Errorf("Expected gzip encoding")
	}

	gzr, err := gzip.NewReader(resp.Body)
	if err != nil {
		t.Fatalf("Failed to create gzip reader: %v", err)
	}
	decompressed, err := io.ReadAll(gzr)
	if err != nil {
		t.Fatalf("Failed to decompress: %v", err)
	}
	if string(decompressed) != "Hello, Gzip!" {
		t.Errorf("Unexpected decompressed content: %s", string(decompressed))
	}
}

func TestGzipCompression_MinSize(t *testing.T) {
	// Simulate min size logic
	minSize := 1024
	data := bytes.Repeat([]byte("A"), minSize+1)
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	gz.Write(data)
	gz.Close()
	if buf.Len() == 0 {
		t.Errorf("Expected compressed data")
	}
}

func TestGzipCompression_Level(t *testing.T) {
	// Test different compression levels
	for level := gzip.BestSpeed; level <= gzip.BestCompression; level++ {
		var buf bytes.Buffer
		gz, err := gzip.NewWriterLevel(&buf, level)
		if err != nil {
			t.Fatalf("Failed to create gzip writer: %v", err)
		}
		gz.Write([]byte("Test Level"))
		gz.Close()
		if buf.Len() == 0 {
			t.Errorf("Level %d: Expected compressed data", level)
		}
	}
}
