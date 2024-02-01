package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

// The watcher struct
type Watcher struct {
	Config
	*fsnotify.Watcher

	eventCh           chan string
	watchingProcessCh chan bool
	exitProgramCh     chan bool
	exitProcessCh     chan bool
}

// Creates a new watcher with a config object
func newWatcher(config Config) (*Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &Watcher{
		Config:            config,
		Watcher:           watcher,
		eventCh:           make(chan string),
		watchingProcessCh: make(chan bool),
		exitProgramCh:     make(chan bool),
		exitProcessCh:     make(chan bool),
	}, nil
}

// Starts watching files
func (w *Watcher) Start() error {
	defer w.Watcher.Close()

	err := w.Add(w.RootDirectory)
	if err == nil {
		err = filepath.Walk(w.RootDirectory, w.watchDirectory)
	}

	if err != nil {
		return err
	}

	fmt.Println(">>> Starting reload server <<<")

	go w.runCmd(w.RunCmd)

	fmt.Println(">>> Listening for events <<<")

	doneCh := make(chan bool)
	go func() {
		for {
			select {
			case <-w.exitProgramCh:
				doneCh <- true
				return

			case path := <-w.Events:
				fileInfo, err := os.Stat(path.Name)

				if err == nil && fileInfo.IsDir() {
					err = filepath.Walk(w.RootDirectory, w.watchDirectory)
				}

				if err != nil {
					fmt.Printf(">>> Error while trying to watch new directory: %s <<<\n", err)
					continue
				}

				w.stopRunningProcess()

				if !fileInfo.IsDir() && w.shouldWatchFile(fileInfo.Name()) {
					fmt.Printf(">>> File %s modified. Reloading <<<\n", fileInfo.Name())
					go w.runCmd(w.getRunCmd(fileInfo.Name()))
				}
			}

			time.Sleep(time.Duration(w.Delay))
			w.flushEvents()
		}
	}()

	<-doneCh
	return nil
}

// Flushes all events from the fsnotify event channel
func (w *Watcher) flushEvents() {
	for {
		select {
		case <-w.Events:
			time.Sleep(time.Duration(float32(time.Second) * 0.01))
		default:
			return
		}
	}
}

// Signals to the watcher to stop
func (w *Watcher) stop() {
	fmt.Println(">>> Stopping live reload server <<<")

	w.stopRunningProcess()
	w.exitProgramCh <- true
}

func (w *Watcher) stopRunningProcess() {
	select {
	case <-w.watchingProcessCh:
		w.exitProcessCh <- true
	default:
	}
}

// Runs a new command
func (w *Watcher) runCmd(cmd string) error {
	c, stdout, stderr, err := startCmd(cmd)
	if err != nil {
		return err
	}

	_, _ = io.Copy(os.Stdout, stdout)
	_, _ = io.Copy(os.Stderr, stderr)

	go func() {
		w.watchingProcessCh <- true
	}()

	processFinishedCh := make(chan error)
	go func() {
		processFinishedCh <- c.Wait()
	}()

	select {
	case <-w.exitProcessCh:
		err = c.Process.Kill()
	case err = <-processFinishedCh:
	}

	stdout.Close()
	stderr.Close()

	select {
	case <-w.watchingProcessCh:
	default:
	}

	fmt.Println(">>> Process finished <<<")
	return err
}

// Adds a directory to the watcher object
func (w *Watcher) watchDirectory(directory string, info os.FileInfo, err error) error {
	if err == nil && w.shouldWatchDir(directory) {
		return w.Add(directory)
	}

	return err
}
