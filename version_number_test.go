package tgxlib

import "testing"

func TestStringUnpack(t *testing.T) {
	verstr := "1.0.1"
	version, err := UnpackVersionFromString(verstr)
	if err != nil {
		t.Errorf("Fail with error %v", err)
		return
	}
	if version.major != 1 || version.minor != 0 || version.patch != 1 || version.subpatch != 0 {
		t.Errorf("Want 1.0.1.0; got %v", version)
	}
}

func TestPack(t *testing.T) {
	version := VersionNumber{
		major: 1,
		minor: 0,
		patch: 1,
	}
	packed := version.pack()
	if packed != 1000100 {
		t.Errorf("Want 1000100; got %d", packed)
	}
}

func TestString(t *testing.T) {
	{
		version := VersionNumber{
			major: 1,
			minor: 2,
		}
		verstr := version.String()
		if verstr != "1.2.0" {
			t.Errorf("Want 1.2.0; got %s", verstr)
		}
	}
	{
		version := VersionNumber{
			major: 1,
			subpatch: 3,
		}
		verstr := version.String()
		if verstr != "1.0.0.3" {
			t.Errorf("Want 1.0.0.3; got %s", verstr)
		}
	}
}
