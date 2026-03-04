package main

import (
	"os"

	"github.com/ninjops/ninjops/internal/app"
)

func main() {
	if err := app.Execute(); err != nil {
		os.Exit(1)
	}
}
