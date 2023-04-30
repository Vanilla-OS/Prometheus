package tests

import (
	"testing"

	"github.com/vanilla-os/prometheus"
)

func TestNewPrometheus(t *testing.T) {
	pmt, err := prometheus.NewPrometheus("storage", "vfs")
	if err != nil {
		t.Fatalf("error creating Prometheus instance: %v", err)
	}

	if pmt == nil {
		t.Fatal("prometheus instance is nil")
	}
}
