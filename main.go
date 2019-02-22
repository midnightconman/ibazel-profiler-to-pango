package main

import (
	"bufio"
	"encoding/json"
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
		log.Fatalf("File open error: %#v", err)
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("Error creating watcher: %#v", err)
	}
	defer watcher.Close()
	err = watcher.Add(filename)
	if err != nil {
		log.Fatalf("Error adding watcher: %#v", err)
	}

	r := bufio.NewReader(f)
	for {
		by, err := r.ReadBytes('\n')
		if err != nil && err != io.EOF {
			return err
		}
		if err != io.EOF {
			log.Infof("Handling event: %s", string(by))
			currentEvent, err := handle(by)
			if err != nil {
				return err
			}
			err = writeFile(currentEvent)
			if err != nil {
				log.Errorf("writeFile error: %+v", err)
			}
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

func handle(b []byte) (string, error) {
	e := Event{}
	err := json.Unmarshal(b, &e)
	if err != nil {
		log.Errorf("json unmarshal error: %#v", err)
		return currentEvent, nil
	}

	switch e.Type {
	case "BUILD_DONE", "TEST_DONE":
		currentEvent = "!Ybg0xff2a6f78Y!" + e.Type
	case "BUILD_FAILED", "TEST_FAILED":
		currentEvent = "!Ybg0xff8b0500Y!" + e.Type
	case "BUILD_START", "TEST_START":
		currentEvent = "!Ybg0xff404040Y!" + e.Type
	}
	return currentEvent, nil
}

func writeFile(s string) error {
	f, err := os.OpenFile(*outputFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("OpenFile error: %+v", err)
	}
	if _, err := f.Write([]byte(s)); err != nil {
		return fmt.Errorf("File Write error: %+v", err)
	}
	f.Sync()
	if err := f.Close(); err != nil {
		return fmt.Errorf("File Close error: %+x", err)
	}
	return nil
}

var (
	home         = os.Getenv("HOME")
	file         = flag.String("file", home+"/.cache/ibazel-profile.json", "The name and path of the file to follow.")
	outputFile   = flag.String("output-file", home+"/.cache/ibazel-event", "The name and path of the output file.")
	currentEvent = "!Ybg0xff000000Y!NO_DATA"
)

func main() {
	flag.Parse()
	log.Infof("Starting to follow %s", *file)
	follow(*file)
}
