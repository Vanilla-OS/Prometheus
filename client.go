package prometheus

/*	License: GPLv3
	Authors:
		Mirko Brombin <mirko@fabricators.ltd>
		Vanilla OS Contributors <https://github.com/vanilla-os/>
	Copyright: 2023
	Description:
		Prometheus is a simple and accessible library for pulling and mounting
		container images. It is designed to be used as a dependency in ABRoot
		and Albius.
*/

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/containers/buildah/define"
	"github.com/containers/buildah/imagebuildah"
	"github.com/containers/image/v5/copy"
	"github.com/containers/image/v5/manifest"
	"github.com/containers/image/v5/signature"
	"github.com/containers/image/v5/storage"
	"github.com/containers/image/v5/transports/alltransports"
	"github.com/containers/image/v5/types"
	cstorage "github.com/containers/storage"
	digest "github.com/opencontainers/go-digest"
)

/* NewPrometheus creates a new Prometheus instance, note that currently
 * Prometheus only works with custom stores, so you need to pass the
 * root graphDriverName to create a new one.
 */
func NewPrometheus(root, graphDriverName string, maxParallelDownloads uint) (*Prometheus, error) {
	var err error

	root = filepath.Clean(root)
	if _, err := os.Stat(root); os.IsNotExist(err) {
		err = os.MkdirAll(root, 0755)
		if err != nil {
			return nil, err
		}
	}

	runRoot := filepath.Join(root, "run")
	graphRoot := filepath.Join(root, "graph")

	store, err := cstorage.GetStore(cstorage.StoreOptions{
		RunRoot:         runRoot,
		GraphRoot:       graphRoot,
		GraphDriverName: graphDriverName,
	})
	if err != nil {
		return nil, err
	}

	return &Prometheus{
		Store: store,
		Config: PrometheusConfig{
			Root:                 root,
			GraphDriverName:      graphDriverName,
			MaxParallelDownloads: maxParallelDownloads,
		},
	}, nil
}

// PullImage pulls an image from a remote registry and stores it in the
// Prometheus store. It returns the manifest of the pulled image and an
// error if any. Note that the 'docker://' prefix is automatically added
// to the imageName to make it compatible with the alltransports.ParseImageName
// method.
func (p *Prometheus) PullImage(imageName, dstName string) (*OciManifest, digest.Digest, error) {
	progressCh := make(chan types.ProgressProperties)
	manifestCh := make(chan OciManifest)
	errorCh := make(chan error)

	defer close(progressCh)
	defer close(manifestCh)
	defer close(errorCh)

	manifest, manifestDigest, err := p.PullManifestOnly(imageName)
	if err != nil {
		return nil, "", fmt.Errorf("failed to pull manifest %w", err)
	}

	// +1 to account for manifest
	layersAll := len(manifest.LayerInfos()) + 1
	layersDone := 0

	err = p.pullImage(imageName, dstName, progressCh, manifestCh, errorCh)
	if err != nil {
		return nil, "", err
	}
	for {
		select {
		case report := <-progressCh:
			digestShort := report.Artifact.Digest.Encoded()[:12]
			switch report.Event {
			case types.ProgressEventNewArtifact:
				fmt.Printf("[%v/%v] %s: %s\n", layersDone, layersAll, digestShort, "new")
			case types.ProgressEventDone, types.ProgressEventSkipped:
				layersDone += 1
				fmt.Printf("[%v/%v] %s: %s\n", layersDone, layersAll, digestShort, "done")
			case types.ProgressEventRead:
				percentDone := 100 * float64(report.Offset) / float64(report.Artifact.Size)
				fmt.Printf("[%v/%v] %s: %v%%\n", layersDone, layersAll, digestShort, int(percentDone))
			}
		case manifest := <-manifestCh:
			return &manifest, manifestDigest, nil
		case err := <-errorCh:
			return nil, "", err
		}
	}
}

// PullImageAsync does the same thing as PullImage, but returns right
// after starting the pull process. The user can track progress in the
// background by reading from the `progressCh` channel, which contains
// information about the current blob and its progress. When the pull
// process is done, the image's manifest will be sent via the `manifestCh`
// channel, which indicates the process is done.
//
// NOTE: The user is responsible for closing both channels once the operation
// completes.
func (p *Prometheus) PullImageAsync(imageName, dstName string, progressCh chan types.ProgressProperties, manifestCh chan OciManifest, errorCh chan error) error {
	err := p.pullImage(imageName, dstName, progressCh, manifestCh, errorCh)
	return err
}

func (p *Prometheus) pullImage(imageName, dstName string, progressCh chan types.ProgressProperties, manifestCh chan OciManifest, errorCh chan error) error {
	srcRef, err := alltransports.ParseImageName(fmt.Sprintf("docker://%s", imageName))
	if err != nil {
		return err
	}

	destRef, err := storage.Transport.ParseStoreReference(p.Store, dstName)
	if err != nil {
		return err
	}

	systemCtx := &types.SystemContext{}
	policy, err := signature.DefaultPolicy(systemCtx)
	if err != nil {
		return err
	}

	policyCtx, err := signature.NewPolicyContext(policy)
	if err != nil {
		return err
	}

	duration, err := time.ParseDuration("100ms")
	if err != nil {
		return err
	}

	go func() {
		pulledManifestBytes, err := copy.Image(
			context.Background(),
			policyCtx,
			destRef,
			srcRef,
			&copy.Options{
				MaxParallelDownloads: p.Config.MaxParallelDownloads,
				ProgressInterval:     duration,
				Progress:             progressCh,
			},
		)
		if err != nil {
			errorCh <- err
			return
		}

		var manifest OciManifest
		err = json.Unmarshal(pulledManifestBytes, &manifest)
		if err != nil {
			errorCh <- err
			return
		}

		manifestCh <- manifest
	}()

	return nil
}

/* GetImageById returns an image from the Prometheus store by its ID. */
func (p *Prometheus) GetImageById(id string) (cstorage.Image, error) {
	images, err := p.Store.Images()
	if err != nil {
		return cstorage.Image{}, err
	}

	for _, img := range images {
		if img.ID == id {
			return img, nil
		}
	}

	err = cstorage.ErrImageUnknown
	return cstorage.Image{}, err
}

/* DoesImageExist checks if an image exists in the Prometheus store by its ID.
 * It returns a boolean indicating if the image exists and an error if any. */
func (p *Prometheus) DoesImageExist(id string) (bool, error) {
	image, err := p.GetImageById(id)
	if err != nil {
		return false, err
	}

	if image.ID == id {
		return true, nil
	}

	return false, nil
}

/* MountImage mounts an image from the Prometheus store by its main layer
 * digest. It returns the mount path and an error if any. */
func (p *Prometheus) MountImage(layerId string) (string, error) {
	mountPath, err := p.Store.Mount(layerId, "")
	if err != nil {
		return "", err
	}

	return mountPath, nil
}

/* UnMountImage unmounts an image from the Prometheus store by its main layer
 * digest. It returns a boolean indicating if the unmount was successful and
 * an error if any. */
func (p *Prometheus) UnMountImage(layerId string, force bool) (bool, error) {
	res, err := p.Store.Unmount(layerId, force)
	if err != nil {
		return res, err
	}

	return res, nil
}

/* BuildContainerFile builds a dockerfile and returns the manifest of the built
 * image and an error if any. */
func (p *Prometheus) BuildContainerFile(dockerfilePath string, imageName string) (cstorage.Image, error) {
	id, _, err := imagebuildah.BuildDockerfiles(
		context.Background(),
		p.Store,
		define.BuildOptions{
			Output: imageName,
		},
		dockerfilePath,
	)
	if err != nil {
		return cstorage.Image{}, err
	}

	image, err := p.GetImageById(id)
	if err != nil {
		return cstorage.Image{}, err
	}

	return image, nil
}

func (p *Prometheus) PullManifestOnly(imageName string) (manifest.Manifest, digest.Digest, error) {
	systemCtx := &types.SystemContext{}
	ctx := context.Background()

	var outputManifest manifest.Manifest

	srcRef, err := alltransports.ParseImageName(fmt.Sprintf("docker://%s", imageName))
	if err != nil {
		return outputManifest, "", fmt.Errorf("failed to parse image name: %w", err)
	}

	source, err := srcRef.NewImageSource(ctx, systemCtx)
	if err != nil {
		return outputManifest, "", fmt.Errorf("failed to create image source: %w", err)
	}
	defer source.Close()

	manRaw, manMime, err := source.GetManifest(ctx, nil)
	if err != nil {
		return outputManifest, "", fmt.Errorf("failed to fetch manifest: %w", err)
	}

	if manifest.MIMETypeIsMultiImage(manMime) {
		list, err := manifest.ListFromBlob(manRaw, manMime)
		if err != nil {
			return outputManifest, "", fmt.Errorf("failed to parse manifest list: %w", err)
		}

		instanceDigest, err := list.ChooseInstance(systemCtx)
		if err != nil {
			return outputManifest, "", fmt.Errorf("failed to select platform instance: %w", err)
		}

		manRaw, manMime, err = source.GetManifest(ctx, &instanceDigest)
		if err != nil {
			return outputManifest, "", fmt.Errorf("failed to fetch platform manifest: %w", err)
		}
	}

	outputManifest, err = manifest.FromBlob(manRaw, manMime)
	if err != nil {
		return outputManifest, "", fmt.Errorf("failed to parse platform specific manifest: %w", err)
	}

	var manifestDigest digest.Digest
	manifestDigest, err = manifest.Digest(manRaw)
	if err != nil {
		return outputManifest, "", fmt.Errorf("failed to fetch digest of manifest: %w", err)
	}

	return outputManifest, manifestDigest, nil
}
