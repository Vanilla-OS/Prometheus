package main

import (
	"fmt"

	"github.com/containers/storage/pkg/reexec"
	"github.com/vanilla-os/prometheus"
)

func main() {
	if reexec.Init() {
		return
	}

	// -----------------------------
	fmt.Println("Building image from a Containerfile...")

	pmt, err := prometheus.NewPrometheus("storage", "overlay", 5)
	if err != nil {
		panic(err)
	}

	image, err := pmt.BuildContainerFile("Containerfile", "example")
	if err != nil {
		panic(err)
	}

	fmt.Printf("Image built: %s\n", image.ID)

	// -----------------------------

	fmt.Println("Mounting top layer...")

	mountPath, err := pmt.MountImage(image.TopLayer)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Mounted at: %s\n", mountPath)
}
