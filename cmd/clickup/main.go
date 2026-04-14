package main

import (
	"os"

	"github.com/toppynl/clickup-cli/internal/app"
)

func main() {
	code := app.Run()
	os.Exit(code)
}
