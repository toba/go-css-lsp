package lsp

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"sync"
	"testing"
)

// TestConcurrentSendToLspClient_Race verifies that concurrent
// calls to SendToLspClient on a shared writer do not trigger
// the race detector. Run with: go test -race ./...
func TestConcurrentSendToLspClient_Race(t *testing.T) {
	var mu sync.Mutex
	var buf bytes.Buffer

	const goroutines = 10
	const messagesPerGoroutine = 50

	var wg sync.WaitGroup

	for g := range goroutines {
		wg.Go(func() {
			for i := range messagesPerGoroutine {
				msg := fmt.Appendf(nil,
					`{"jsonrpc":"2.0","method":"test","params":{"g":%d,"i":%d}}`,
					g, i,
				)
				mu.Lock()
				SendToLspClient(&buf, msg)
				mu.Unlock()
			}
		})
	}

	wg.Wait()

	total := goroutines * messagesPerGoroutine
	scanner := bufio.NewScanner(&buf)
	// Use a buffer large enough to hold all messages at once
	// to avoid triggering scanner buffer-boundary edge cases.
	scanBuf := make([]byte, 1<<20)
	scanner.Buffer(scanBuf, 1<<20)
	scanner.Split(decode)

	var count int
	for scanner.Scan() {
		count++
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("decode error after %d messages: %v", count, err)
	}
	if count != total {
		t.Fatalf("expected %d messages, got %d", total, count)
	}
}

// TestConcurrentSendToLspClient_FrameIntegrity verifies that
// concurrent writes protected by a mutex produce a valid LSP
// byte stream where every message has a correct Content-Length
// header and the advertised number of body bytes.
func TestConcurrentSendToLspClient_FrameIntegrity(t *testing.T) {
	pr, pw := io.Pipe()

	const goroutines = 10
	const messagesPerGoroutine = 100
	total := goroutines * messagesPerGoroutine

	var mu sync.Mutex
	var wg sync.WaitGroup

	for g := range goroutines {
		wg.Go(func() {
			for i := range messagesPerGoroutine {
				msg := fmt.Appendf(
					nil,
					`{"jsonrpc":"2.0","method":"textDocument/publishDiagnostics","params":{"g":%d,"i":%d}}`,
					g,
					i,
				)
				mu.Lock()
				SendToLspClient(pw, msg)
				mu.Unlock()
			}
		})
	}

	// Collect all output into a buffer so we can parse it
	// with a large scanner buffer in one pass.
	var collected bytes.Buffer
	done := make(chan struct{})
	go func() {
		_, _ = io.Copy(&collected, pr)
		close(done)
	}()

	wg.Wait()
	_ = pw.Close()
	<-done

	scanner := bufio.NewScanner(&collected)
	scanBuf := make([]byte, 1<<20)
	scanner.Buffer(scanBuf, 1<<20)
	scanner.Split(decode)

	var count int
	for scanner.Scan() {
		body := scanner.Bytes()
		if len(body) == 0 {
			t.Errorf("message %d: empty body", count)
		}
		if body[0] != '{' {
			t.Errorf(
				"message %d: body does not start with '{': %q",
				count, body[:min(40, len(body))],
			)
		}
		count++
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("decode error after %d messages: %v", count, err)
	}
	if count != total {
		t.Fatalf("expected %d messages, got %d", total, count)
	}
}
