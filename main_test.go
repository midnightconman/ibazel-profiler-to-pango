package main

import (
	"testing"
)

func TestHandleSimple(t *testing.T) {
	o, err := handle([]byte(`{"type":"TEST_DONE"}`))
	if err != nil {
		t.Errorf("%+v", err)
	}
	if o != "!Ybg0xff2a6f78Y!TEST_DONE" {
		t.Error("handle output wasn't as expected.")
	}
}
