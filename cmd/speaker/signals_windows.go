//go:build windows

package speaker

import "os"

func shutdownSignals() []os.Signal {
	return []os.Signal{os.Interrupt}
}
