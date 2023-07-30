package tgxlib

import (
	"fmt"
	"strings"
	"strconv"
)

type VersionNumber struct {
	major uint
	minor uint
	patch uint
}

func unpackVersionFromU32(packedversion uint32) VersionNumber {
	var version VersionNumber
	version.patch = uint(packedversion) % 100
	packedversion /= 100
	version.minor = uint(packedversion) % 100
	packedversion /= 100
	version.major = uint(packedversion) % 100
	return version
}

func UnpackVersionFromString(packedversion string) (VersionNumber, error) {
	var version VersionNumber
	var tmp int
	var err error
	parts := strings.Split(packedversion, ".")
	tmp, err = strconv.Atoi(parts[0])
	if err != nil {
		return VersionNumber{}, err
	}
	version.major = uint(tmp)
	tmp, err = strconv.Atoi(parts[1])
	if err != nil {
		return VersionNumber{}, err
	}
	version.minor = uint(tmp)
	tmp, err = strconv.Atoi(parts[2])
	if err != nil {
		return VersionNumber{}, err
	}
	version.patch = uint(tmp)

	return version, nil
}

func (version VersionNumber) pack() uint32 {
	return uint32((version.major * 10000) + (version.minor * 100) + version.patch)
}

func (version VersionNumber) String() string {
	return fmt.Sprintf("%d.%d.%d", version.major, version.minor, version.patch)
}
