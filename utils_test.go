package tgxlib

import (
	"testing"
)

func TestGenIdentifier(t *testing.T) {
	id := genIdentifier("Data\\ModInfo.ini")
	if id != 0x094f736c {
		t.Errorf("Got 0x%08x; want 0x094f736c", id)
	}
}

func TestCStrLen(t *testing.T) {
	buf := [12]byte{30, 30, 30, 30, 30, 0, 0, 0, 0, 0, 0, 0}
	length := cStrLen(buf[:])
	if length != 5 {
		t.Errorf("Got %d; want 5", length)
	}
}

func TestGetSliceSegment(t *testing.T) {
	buf := make([]uint, 40)
	buf[20] = 32
	slice := getSliceSegment(buf, 10, 2)
	if len(slice) != 10 {
		t.Errorf("len(slice) = %d; want 10", len(slice))
	}
	if slice[0] != 32 {
		t.Errorf("slice[0] = %d; want 32", slice[0])
	}
}
