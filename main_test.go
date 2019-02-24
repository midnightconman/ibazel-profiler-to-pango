package main

import (
	"testing"
)

func TestHandleSimple(t *testing.T) {
	color := "ff2a6f78"
	doneColor = &color
	err := handle([]byte(`{"type":"TEST_DONE"}`))
	if err != nil {
		t.Errorf("%+v", err)
	}
	if currentEvent != "!Ybg0xff2a6f78Y!TEST_DONE" {
		t.Errorf("handle output wasn't as expected. %+v", currentEvent)
	}
}
