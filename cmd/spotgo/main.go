/*
Copyright © 2024 Esteban Bermudez Aguirre <esteban@bermudezaguirre.com>
*/
package main

import (
	"github.com/Esteban-Bermudez/spotgo/cmd/root"
	// This ensures the init commands in the sub commands run
	_ "github.com/Esteban-Bermudez/spotgo/cmd/connect"
	_ "github.com/Esteban-Bermudez/spotgo/cmd/daemon"
	_ "github.com/Esteban-Bermudez/spotgo/cmd/devices"
	_ "github.com/Esteban-Bermudez/spotgo/cmd/player"
	_ "github.com/Esteban-Bermudez/spotgo/cmd/play"
	_ "github.com/Esteban-Bermudez/spotgo/cmd/queue"
	_ "github.com/Esteban-Bermudez/spotgo/cmd/search"
)

func main() {
	root.Execute()
}
