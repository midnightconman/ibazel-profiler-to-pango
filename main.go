package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

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
			err := handle(by)
			if err != nil {
				return err
			}
			switch strings.ToLower(*outputMode) {
			default:
				err = writeFile(currentEvent)
				if err != nil {
					log.Errorf("writeFile error: %+v", err)
				}
			case "command":
				cmd := exec.Command("sh", "-c", currentCommand)
				if _, err := cmd.CombinedOutput(); err != nil {
					log.Errorf("Error running command: %+v", err)
				}
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

func handle(b []byte) error {
	e := Event{}
	err := json.Unmarshal(b, &e)
	if err != nil {
		log.Errorf("json unmarshal error: %#v", err)
		return err
	}

	switch e.Type {
	case "BUILD_DONE", "TEST_DONE":
		currentEvent = "!Ybg0x"+*doneColor+"Y!" + e.Type
		currentCommand = *doneCommand
	case "BUILD_FAILED", "TEST_FAILED":
		currentEvent = "!Ybg0x"+*failedColor+"Y!" + e.Type
		currentCommand = *failedCommand
	case "BUILD_START", "TEST_START":
		currentEvent = "!Ybg0x"+*startColor+"Y!" + e.Type
		currentCommand = *startCommand
	}
	return nil
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
	home          = os.Getenv("HOME")
	file          = flag.String("file", home+"/.cache/ibazel-profile.json", "The name and path of the file to follow.")
	outputMode    = flag.String("output-mode", "file", "If you would like to use file or command mode.")
	outputFile    = flag.String("output-file", home+"/.cache/ibazel-event", "The name and path of the output file.")
	doneColor     = flag.String("done-color", "ff2a6f78", "The hexidecimal color to output on *-DONE events. Including alpha, ie. ff000000")
	failedColor   = flag.String("failed-color", "ff8b0500", "The hexidecimal color to output on *-FAILED events. Including alpha, ie. ff000000")
	startColor    = flag.String("start-color", "ff404040", "The hexidecimal color to output on *-START events. Including alpha, ie. ff000000")
	// Command flags
	doneCommand   = flag.String("done-command", "echo", "The command to execute on *-DONE events")
	failedCommand = flag.String("failed-command", "echo", "The command to execute on *-FAILED events")
	startCommand  = flag.String("start-command", "echo", "The command to execute on *-START events")
	// (TODO): move these variables out into an event struct
	currentEvent   = "!Ybg0xff000000Y!NO_DATA"
	currentCommand = "echo"
)

func main() {
	flag.Parse()
	log.Infof("Starting to follow %s", *file)
	follow(*file)
}
