package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// The Flags struct
type Flags struct {
	configFile string
}

// Creates a new Flags struct with default values
func newDefaultFlags() Flags {
	return Flags{
		configFile: "gowatch.toml",
	}
}

// Tries parsing system arguments into a flags struct choosing to use default
// values when the user does not provide them
func parseFlags(args []string, flags Flags) Flags {
	flag.StringVar(&flags.configFile, "f", flags.configFile, "The configuration file")
	return flags
}

func main() {
	// Create default program flags and then try to parse any system flags
	flags := newDefaultFlags()
	if len(os.Args) > 2 {
		flags = parseFlags(os.Args, flags)
	}

	// Create a Config object from the provided toml file
	content, err := os.ReadFile(flags.configFile)
	if err != nil {
		fmt.Printf(">>> Error trying to read config file: %s <<<\n", err)
	}

	config, err := newConfig(content)
	if err != nil {
		fmt.Printf(">>> Error trying to parse config file: %s <<<\n", err)
	}

	// Create a new watcher and start it
	interruptCh := make(chan os.Signal)
	signal.Notify(interruptCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	watcher, err := newWatcher(config)
	if err == nil {
		go func() {
			<-interruptCh
			watcher.stop()
		}()

		err = watcher.Start()
	}

	if err != nil {
		fmt.Printf(">>> failed to start watching files: %s <<<\n", err)
		return
	}
}
