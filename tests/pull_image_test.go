package tests

import (
	"fmt"
	"os"
	"testing"

	"go.podman.io/image/v5/types"
	"go.podman.io/storage/pkg/reexec"
	"github.com/vanilla-os/prometheus"
)

var pmt *prometheus.Prometheus

func TestMain(m *testing.M) {
	if reexec.Init() {
		return
	}

	var err error
	pmt, err = prometheus.NewPrometheus("storage", "vfs", 5)
	if err != nil {
		panic("error creating Prometheus instance: " + err.Error())
	}

	if pmt == nil {
		panic("prometheus instance is nil")
	}

	status := m.Run()
	os.Exit(status)
}

func TestPullImage(t *testing.T) {
	image, digest, err := pmt.PullImage("docker.io/library/alpine:latest", "my-alpine")
	if err != nil {
		t.Fatalf("error pulling image: %v", err)
	}

	if image == nil {
		t.Fatal("image is nil")
	}

	if digest == "" {
		t.Fatal("image digest is empty")
	}

	if image.Config.Digest == "" {
		t.Fatal("image config digest is empty")
	}

	if image.Config.Size == 0 {
		t.Fatal("image config size is 0")
	}

	if image.Config.MediaType == "" {
		t.Fatal("image config media type is empty")
	}

	if len(image.Layers) == 0 {
		t.Fatal("image layers is empty")
	}

	for _, layer := range image.Layers {
		if layer.Digest == "" {
			t.Fatal("image layer digest is empty")
		}

		if layer.Size == 0 {
			t.Fatal("image layer size is 0")
		}

		if layer.MediaType == "" {
			t.Fatal("image layer media type is empty")
		}
	}
}

func TestPullImageAsync(t *testing.T) {
	progressCh := make(chan types.ProgressProperties)
	manifestCh := make(chan prometheus.OciManifest)
	errorCh := make(chan error)

	defer close(progressCh)
	defer close(manifestCh)

	err := pmt.PullImageAsync("docker.io/library/alpine:latest", "my-alpine", progressCh, manifestCh, errorCh)
	if err != nil {
		t.Fatalf("error pulling image: %v", err)
	}

	for {
		select {
		case report := <-progressCh:
			fmt.Printf("%s: %v/%v\n", report.Artifact.Digest.Encoded()[:12], report.Offset, report.Artifact.Size)
		case manifest := <-manifestCh:
			fmt.Printf("Got manifest: %v\n", manifest)
			return
		case err := <-errorCh:
			t.Error(err)
		}
	}
}
