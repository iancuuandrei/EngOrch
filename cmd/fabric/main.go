// Command fabric runs Fabric's native command-line interface.
package main

import (
	"os"

	"harness.local/engorch/internal/entrypoint"
)

func main() {
	entrypoint.Run("fabric", os.Args[1:], nil)
}
