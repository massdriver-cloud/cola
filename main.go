package main

import (
	"os"

	"github.com/massdriver-cloud/cola/cmd"

	"github.com/lightstep/otel-launcher-go/launcher"
)

func main() {
	exitCode := 0
	defer func() { os.Exit(exitCode) }()

	// Setup Tracing
	if os.Getenv("LS_ACCESS_TOKEN") != "" {
		otelLauncher := launcher.ConfigureOpentelemetry(
			launcher.WithServiceName("cola"),
			launcher.WithPropagators([]string{"tracecontext"}),
		)
		defer otelLauncher.Shutdown()
	}

	// Run application
	if err := cmd.Execute(); err != nil {
		exitCode = 1
		return
	}
}
