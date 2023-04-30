package prometheus // import "github.com/vanilla-os/prometheus"


VARIABLES

var version = "0.1.4"

FUNCTIONS

func main()

TYPES

type OciManifest struct {
	SchemaVersion int                 `json:"schemaVersion"`
	MediaType     string              `json:"mediaType"`
	Config        OciManifestConfig   `json:"config"`
	Layers        []OciManifestConfig `json:"layers"`
}

type OciManifestConfig struct {
	MediaType string `json:"mediaType"`
	Size      int    `json:"size"`
	Digest    string `json:"digest"`
}

type Prometheus struct {
	Store cstorage.Store
}

func NewPrometheus(root, graphDriverName string) (*Prometheus, error)
    NewPrometheus creates a new Prometheus instance, note that currently *
    Prometheus only works with custom stores, so you need to pass the * root
    graphDriverName to create a new one.

func (p *Prometheus) BuildContainerFile(dockerfilePath string, imageName string) (cstorage.Image, error)
    BuildContainerFile builds a dockerfile and returns the manifest of the built
    * image and an error if any.

func (p *Prometheus) DoesImageExist(digest string) (bool, error)
    DoesImageExist checks if an image exists in the Prometheus store by its *
    digest. It returns a boolean indicating if the image exists and an error *
    if any.

func (p *Prometheus) GetImageByDigest(digest string) (cstorage.Image, error)
    GetImageByDigest returns an image from the Prometheus store by its digest.

func (p *Prometheus) MountImage(layerId string) (string, error)
    MountImage mounts an image from the Prometheus store by its main layer *
    digest. It returns the mount path and an error if any.

func (p *Prometheus) PullImage(imageName string, dstName string) (*OciManifest, error)
    PullImage pulls an image from a remote registry and stores it in the *
    Prometheus store. It returns the manifest of the pulled image and an *
    error if any. Note that the 'docker://' prefix is automatically added * to
    the imageName to make it compatible with the alltransports.ParseImageName *
    method.

func (p *Prometheus) UnMountImage(layerId string, force bool) (bool, error)
    UnMountImage unmounts an image from the Prometheus store by its main layer *
    digest. It returns a boolean indicating if the unmount was successful and *
    an error if any.

