package core

import (
	"flag"
	"os"
)

type StartupOption func(so *startupOptions)

type startupOptions struct {
	debug bool
}

func DebugMode() StartupOption {
	return func(so *startupOptions) {
		so.debug = true
	}
}

func Start(options ...StartupOption) {
	so := startupOptions{
		debug: false,
	}

	for _, option := range options {
		option(&so)
	}

	if !so.debug {
		so.debug = checkForDebug()
	}

	start(so)
}

func checkForDebug() bool {
	//flag
	debug := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()

	//env var
	if envDebug := os.Getenv("DEBUG"); envDebug == "true" {
		*debug = true
	}

	return *debug
}
