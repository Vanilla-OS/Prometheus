package tests

import (
	"testing"

	"github.com/containers/storage/pkg/reexec"
	"github.com/vanilla-os/prometheus"
)

func TestPullImage(t *testing.T) {
	if reexec.Init() {
		return
	}

	pmt, err := prometheus.NewPrometheus("storage", "vfs")
	if err != nil {
		t.Fatalf("error creating Prometheus instance: %v", err)
	}

	if pmt == nil {
		t.Fatal("prometheus instance is nil")
	}

	image, err := pmt.PullImage("docker.io/library/alpine:latest", "my-alpine")
	if err != nil {
		t.Fatalf("error pulling image: %v", err)
	}

	if image == nil {
		t.Fatal("image is nil")
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
