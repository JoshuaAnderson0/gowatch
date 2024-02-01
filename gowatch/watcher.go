package main

import (
	"fmt"
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
	processFinishedCh chan bool
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
		processFinishedCh: make(chan bool),
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

	programDoneCh := make(chan bool)
	go func() {
		for {
			select {
			case <-w.exitProgramCh:
				programDoneCh <- true

			case path := <-w.Events:
				if w.shouldWatchFile(path.Name) {
					fmt.Println()
					fmt.Printf(">>> File %s modified. Reloading <<<\n", path.Name)
					w.stopRunningProcess()
					go w.runCmd(w.getRunCmd(path.Name))
				}
			}

			w.flushEvents()
		}
	}()

	<-programDoneCh
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
		fmt.Println(">>> Killing process <<<")
		w.exitProcessCh <- true
		<-w.processFinishedCh
	case <-w.processFinishedCh:
	default:
	}
}

// Runs a new command
func (w *Watcher) runCmd(cmd string) {
	fmt.Println()
	fmt.Printf(">>> Starting command: %s <<<\n", cmd)

	go func() {
		w.watchingProcessCh <- true
	}()

	c, err := startCmd(cmd)
	if err != nil {
		fmt.Printf("\n>>> Error when starting process: %s <<<\n", err)
	}

	defer func() {
		fmt.Println(">>> Process finished <<<")

		select {
		case <-w.watchingProcessCh:
		default:
		}

		w.processFinishedCh <- true
	}()

	fmt.Println(">>> Watching new process <<<")
	fmt.Println()

	processFinishCh := make(chan error)
	go func() {
		processFinishCh <- c.Wait()
	}()

	select {
	case <-w.exitProcessCh:
		if err = killCmd(c); err != nil {
			fmt.Printf("\n>>> Error when terminating process: %s <<<\n", err)
		} else {
			fmt.Println(">>> Process interrupted <<<")
		}
		return

	case err = <-processFinishCh:
		if err = killCmd(c); err != nil {
			fmt.Printf("\n>>> Error when terminating process: %s <<<\n", err)
		} else {
			fmt.Println("\n>>> Process finished by it self <<<")
		}
		return
	}
}

// Adds a directory to the watcher object
func (w *Watcher) watchDirectory(directory string, info os.FileInfo, err error) error {
	if err == nil && w.shouldWatchDir(directory) {
		return w.Add(directory)
	}

	return err
}
