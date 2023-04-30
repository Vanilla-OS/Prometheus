package prometheus

import "github.com/docker/docker/pkg/reexec"

var version = "0.1.0"

func main() {

	if reexec.Init() { // needed for subprocesses
		return
	}
}
