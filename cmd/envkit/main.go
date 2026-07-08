package main

import (
	"os"

	"github.com/fusheng/envkit/internal/cli"
	"github.com/fusheng/envkit/internal/detector"
)

func main() {
	detector.InitDetectionEnvironment()
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
