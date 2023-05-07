package tests

import (
	"io/ioutil"
	"testing"

	"github.com/containers/storage/pkg/reexec"
	"github.com/vanilla-os/prometheus"
)

func TestBuildContainerfile(t *testing.T) {
	if reexec.Init() {
		return
	}

	pmt, err := prometheus.NewPrometheus("storage", "vfs", 5)
	if err != nil {
		t.Fatalf("error creating Prometheus instance: %v", err)
	}

	if pmt == nil {
		t.Fatal("prometheus instance is nil")
	}

	containerfile := []byte("FROM alpine:latest")
	err = ioutil.WriteFile("Containerfile", containerfile, 0644)
	if err != nil {
		t.Fatalf("error creating Containerfile: %v", err)
	}

	image, err := pmt.BuildContainerFile(
		"Containerfile",
		"my-alpine-2",
	)
	if err != nil {
		t.Fatalf("error pulling image: %v", err)
	}

	if image.Digest == "" {
		t.Fatal("image is nil")
	}
}
