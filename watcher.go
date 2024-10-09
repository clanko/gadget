package main

import (
	"github.com/clanko/gadget/cmd"
	"github.com/clanko/gadget/config"
	"github.com/fsnotify/fsnotify"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type watcher struct {
	fsWatcher   *fsnotify.Watcher
	onEvent     func()
	config      config.Config
	watchPaths  []string
	pauseEvents bool
	mu          sync.Mutex
}

func NewWatcher(conf config.Config) watcher {
	paths, err := getNestedPaths(conf.Path)
	if err != nil {
		panic(cmd.FormatDanger(err.Error()))
	}

	for i := range conf.IncludeDirs {
		includeDirs, err := getNestedPaths(conf.IncludeDirs[i])
		if err != nil {
			cmd.PrintfDanger("Failed to walk include dir: %v", conf.IncludeDirs[i])
		}

		paths = append(paths, includeDirs...)
	}

	paths = append(paths, conf.IncludeFiles...)

	return watcher{
		config:     conf,
		watchPaths: paths,
	}
}

func (w *watcher) watch() {
	var err error
	w.fsWatcher, err = fsnotify.NewWatcher()
	if err != nil {
		panic(cmd.FormatDanger("Failed to create new watcher: %s", err))
	}
	defer w.fsWatcher.Close()

	// Start listening for events.
	go w.watchLoop()

	// Add all paths
	for _, path := range w.watchPaths {
		w.watchDir(path)
	}

	<-make(chan struct{}) // Block forever
}

func (w *watcher) isExcluded(path string) bool {
	// make sure we're not dealing with some nonsense.
	if path == "" {
		return true
	}

	// todo: exclude prefixes and exts
	//for i := range w.config.ExcludePrefix {
	//	// check if str begins with
	//	print("check " + w.config.ExcludePrefix[i])
	//}
	//
	//for i := range w.config.ExcludeExts {
	//	// check if ends with
	//	print("check " + w.config.ExcludeExts[i])
	//}

	combinedNames := append(w.config.ExcludeDirs, w.config.ExcludeFiles...)

	for i := range combinedNames {
		if combinedNames[i] == path {
			return true
		}
	}

	return false
}

func (w *watcher) watchDir(dir string) {
	if w.isExcluded(dir) == false {
		err := w.fsWatcher.Add(dir)

		if verbose > 0 {
			cmd.PrintfInfo("watching %v", dir)
		}

		if err != nil {
			panic(cmd.FormatDanger("%q: %s", dir, err))
		}
	}
}

func (w *watcher) watchLoop() {
	var (
		wait   = 500 * time.Millisecond
		mu     sync.Mutex
		timers = make(map[string]*time.Timer)
	)

	for {
		select {
		case err, ok := <-w.fsWatcher.Errors:
			if !ok {
				return
			}
			cmd.PrintfDanger(err.Error())

		case e, ok := <-w.fsWatcher.Events:
			if !ok {
				return
			}

			w.mu.Lock()
			isPaused := w.pauseEvents
			w.mu.Unlock()

			if isPaused {
				continue
			}

			if w.isExcluded(e.Name) {
				continue
			}

			if w.isExcluded(e.Name) == false && e.Has(fsnotify.Create) {
				// If a directory was created, walk and watch
				_, err := os.ReadDir(e.Name)
				if err == nil {
					// we got a dir!
					dirStructure, err := getNestedPaths(e.Name)
					if err != nil {
						panic("failed to walk dir path of added directory " + e.Name)
					}

					for i := range dirStructure {
						w.watchDir(dirStructure[i])
					}
				}
			}

			mu.Lock()
			modifiedTimer, ok := timers["fileModified"]
			mu.Unlock()

			if !ok {
				modifiedTimer = time.AfterFunc(math.MaxInt64, func() {
					// something in the event causes a rebuild and sometimes continuously cycles.
					// If we just print something, it only prints once.
					// pausing events seems to have prevented it.
					// todo: come back to this and figure out what actually was wrong
					w.mu.Lock()
					isPaused := w.pauseEvents
					w.mu.Unlock()

					if isPaused == false {
						w.mu.Lock()
						w.pauseEvents = true
						w.mu.Unlock()

						w.onEvent()

						w.mu.Lock()
						w.pauseEvents = false
						w.mu.Unlock()
					}

					mu.Lock()
					delete(timers, "fileModified")
					mu.Unlock()
				})
				modifiedTimer.Stop()

				mu.Lock()
				timers["fileModified"] = modifiedTimer
				mu.Unlock()
			}

			modifiedTimer.Reset(wait)
		}
	}
}

func getNestedPaths(root string) ([]string, error) {
	var dirs []string
	err := filepath.WalkDir(root, func(path string, info os.DirEntry, err error) error {
		// Don't watch hidden files
		if strings.Contains(path, ".") {
			return nil
		}

		if info.IsDir() {
			dirs = append(dirs, path)
		}
		return nil
	})
	return dirs, err
}
