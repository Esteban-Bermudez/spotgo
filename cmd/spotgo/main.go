/*
Copyright © 2024 Esteban Bermudez Aguirre <esteban@bermudezaguirre.com>
*/
package main

import (
	"github.com/Esteban-Bermudez/spotgo/cmd/root"
	// This ensures the init commands in the sub commands run
	_ "github.com/Esteban-Bermudez/spotgo/cmd/connect"
	_ "github.com/Esteban-Bermudez/spotgo/cmd/player"
)

func main() {
	root.Execute()
}
