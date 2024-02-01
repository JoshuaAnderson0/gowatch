package main

import (
	"strings"

	"github.com/BurntSushi/toml"
)

// The config struct
type Config struct {
	WatchConfig   `toml:"watch"`
	CommandConfig `toml:"command"`
	RootDirectory string `toml:"root"` // The root directory the programming is watching in
	RunCmd        string `toml:"cmd"`  // The command to run at startup or during a generic reload
}

// The watch config
type WatchConfig struct {
	IncludeFiles []string `toml:"include_files"` // All the files to watch
	IncludeDirs  []string `toml:"include_dirs"`  // All the directories to watch
	ExcludeFiles []string `toml:"exclude_files"` // All the files to exclude
	ExcludeDirs  []string `toml:"exclude_dirs"`  // All the directories to exclude
}

// The command config
type CommandConfig map[string]struct {
	Cmd string `toml:"cmd"` // The command to run when a specific file type is modified
}

// Creates a new config from the contents of a toml file
func newConfig(rawToml []byte) (Config, error) {
	var config Config
	err := toml.Unmarshal(rawToml, &config)

	return config, err
}

// Checks if a directory should be watched or not.
// If `IncludeDirs` is populated then check against that otherwise check against `ExcludeDirs`
func (c Config) shouldWatchDir(directory string) bool {
	if c.IncludeDirs != nil && len(c.IncludeDirs) > 0 {
		for _, include := range c.IncludeDirs {
			if directory == include {
				return true
			}
		}
		return false
	}

	if c.ExcludeDirs != nil && len(c.ExcludeDirs) > 0 {
		for _, exclude := range c.ExcludeDirs {
			if directory == exclude {
				return false
			}
		}
	}

	return true
}

// Checks if a file should be watched or not.
// If `IncludeFiles` is populated then check against that otherwise check against `ExcludeFiles`
func (c Config) shouldWatchFile(file string) bool {
	if c.IncludeFiles != nil && len(c.IncludeFiles) > 0 {
		for _, include := range c.IncludeFiles {
			if strings.HasSuffix(file, include) {
				return true
			}
		}
		return false
	}

	if c.ExcludeFiles != nil && len(c.ExcludeFiles) > 0 {
		for _, exclude := range c.ExcludeFiles {
			if strings.HasSuffix(file, exclude) {
				return false
			}
		}
	}

	return true
}

// Searches through the commands to try and find a match based on the file type that's being built
func (c Config) getRunCmd(file string) string {
	for name, command := range c.CommandConfig {
		if strings.HasSuffix(file, name) {
			return command.Cmd
		}
	}

	return c.RunCmd
}
