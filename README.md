# gowatch
A file live reload application

Based on the golang Air package by [cosmtrek](https://github.com/cosmtrek/) this golang package works with any type of project, any file extensions and command line instructions to build and run your project's processes when changes are detected. 

Gowatch adds the ability to run different build commands for your project depending on what files have been modified.

By using a `config.toml` file you can modify gowatch to work for you.
```
# Root directory for watching
root = "/path/to/root/directory"

# Command to run at startup or during a generic reload
cmd = "go run main.go"

# Watching configuration
[watch]
  # Files to include
  include_files = [".go", ".ts"]

  # Directories to include
  include_dirs = ["src", "static"]

  # Files to exclude
  exclude_files = [".test.go"]

  # Directories to exclude
  exclude_dirs = ["vendor"]

# Example command for .go files
[command.".go"]
cmd = "go run main.go"

# Example command for .ts files
[command.".ts"]
cmd = "tsc -p ."
```

## Notes:
Currently this program only works for windows
