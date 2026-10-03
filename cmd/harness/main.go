// Command harness runs the local deterministic engineering harness.
package main

import (
	"os"

	"harness.local/engorch/internal/entrypoint"
)

func main() {
	entrypoint.Run("harness", os.Args[1:], nil)
}
