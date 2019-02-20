package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
)

type Event struct {
	Type string `json:"type"`
}

func follow(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		log.Errorf("File open error: %+v", err)
	}
	watcher, _ := fsnotify.NewWatcher()
	defer watcher.Close()
	_ = watcher.Add(filename)

	r := bufio.NewReader(f)
	for {
		by, err := r.ReadBytes('\n')
		if err != nil && err != io.EOF {
			return err
		}
		// handle event
		fmt.Print(string(by))
		if err != io.EOF {
			continue
		}
		if err = waitForChange(watcher); err != nil {
			return err
		}
	}
}

func waitForChange(w *fsnotify.Watcher) error {
	for {
		select {
		case event := <-w.Events:
			if event.Op&fsnotify.Write == fsnotify.Write {
				return nil
			}
		case err := <-w.Errors:
			return err
		}
	}
}

var file = flag.String("file", "profile.json", "The name and path of the file to follow.")

func main() {
	log.Infof("Starting to follow %s", *file)
	follow(*file)
}
