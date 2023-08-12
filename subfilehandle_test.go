package tgxlib

import (
	"testing"
)

func TestReadSubfileOffset(t *testing.T) {
	testHandle := &subfileHandle{
		header: nil,
		fileData: []byte{0, 1, 2, 3, 4, 5, 6},
		readOffset: 0,
	}
	buf := make([]byte, 4)
	count, _ := testHandle.Read(buf)
	if count != 4 {
		t.Errorf("Got count=%d; want 4", count)
	}
	if testHandle.readOffset != 4 {
		t.Errorf("Got offset=%d; want 4", testHandle.readOffset)
	}
}
