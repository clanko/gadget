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
	fsWatching  []string
	pauseEvents bool
	mu          sync.Mutex
	isWatching  bool
}

func newWatcher(conf config.Config) watcher {
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

	w.isWatching = true

	<-make(chan struct{}) // Block forever
}

// watcher should ideally watch all directories, and filter out items by Event.Name
func (w *watcher) isExcluded(path string) bool {
	// make sure we're not dealing with some nonsense.
	if path == "" {
		return true
	}

	// check if there's a tilde, if so, return isExcluded true
	if strings.HasSuffix(path, "~") {
		return true
	}

	// included files will not be excluded
	for _, file := range w.config.IncludeFiles {
		if file == path {
			return false
		}
	}

	// exclude files
	for _, file := range w.config.ExcludeFiles {
		if file == path {
			return true
		}
	}

	// now we've gotten to where the path doesn't directly match an include or excluded file.
	// need to check if path begins with an excluded path,
	// if it does, then it should be excluded, unless it matches higher in included files
	excludeMatchChars := 0
	includeMatchChars := 0
	for _, excludedDir := range w.config.ExcludeDirs {
		if excludedDir == path {
			return true
		}

		if strings.HasPrefix(path, excludedDir) {
			excludeMatchChars = len(excludedDir)
		}
	}

	for _, includeDir := range w.config.IncludeDirs {
		if includeDir == path {
			return false
		}

		if strings.HasPrefix(path, includeDir) {
			includeMatchChars = len(includeDir)
		}
	}

	if excludeMatchChars > includeMatchChars {
		return true
	}

	pathParts := strings.Split(path, "/")
	fileName := pathParts[len(pathParts)-1]

	for _, prefix := range w.config.ExcludePrefix {
		if strings.HasPrefix(fileName, prefix) {
			return true
		}
	}

	for _, suffix := range w.config.ExcludeExts {
		if strings.HasSuffix(fileName, suffix) {
			return true
		}
	}

	return false
}

func (w *watcher) watchDir(dir string) {
	stat, err := os.Stat(dir)
	if err != nil {
		cmd.PrintfDanger("%v", err)

		return
	}

	isWatching := false
	if !stat.IsDir() {
		// if not in fsWatching
		filePath := filepath.Dir(dir)
		for _, watching := range w.fsWatching {
			if watching == filePath {
				isWatching = true
			}
		}

		if !isWatching {
			dir = filepath.Dir(dir)
		}
	}

	if !isWatching {
		err = w.fsWatcher.Add(dir)
		w.fsWatching = append(w.fsWatching, dir)

		if err != nil {
			panic(cmd.FormatDanger("%q: %s", dir, err))
		}
	}

	if verbose > 0 {
		if w.isExcluded(dir) {
			cmd.PrintfWarning("skipping %v", dir)
		} else {
			cmd.PrintfInfo("watching %v", dir)
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

			if e.Has(fsnotify.Create) {
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

func (w *watcher) endWatch() {
	if w.isWatching {
		err := w.fsWatcher.Close()
		if err != nil {
			cmd.PrintfDanger("%v", err)

			return
		}

		w.isWatching = false
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
