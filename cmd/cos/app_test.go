package main

import (
	"bytes"
	"testing"
)

func TestRunRejectsUnknownCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := run([]string{"unknown"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error")
	}
}
