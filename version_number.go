package tgxlib

import (
	"fmt"
	"strconv"
	"strings"
)

type VersionNumber struct {
	major uint
	minor uint
	patch uint
	subpatch uint
}

func unpackVersionFromU32(packedversion uint32) VersionNumber {
	var version VersionNumber
	version.subpatch = uint(packedversion) % 100
	packedversion /= 100
	version.patch = uint(packedversion) % 100
	packedversion /= 100
	version.minor = uint(packedversion) % 100
	packedversion /= 100
	version.major = uint(packedversion) % 100
	return version
}

func UnpackVersionFromString(packedversion string) (VersionNumber, error) {
	var version VersionNumber
	//var tmp int
	//var err error
	parts := strings.Split(packedversion, ".")
	/*tmp, err = strconv.Atoi(parts[0])
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
	version.patch = uint(tmp)*/
	for ix, subversion := range []*uint{&version.major, &version.minor, &version.patch, &version.subpatch} {
		if ix >= len(parts) {
			break
		}
		tmp, err := strconv.Atoi(parts[ix])
		if err != nil {
			return VersionNumber{}, err
		}
		*subversion = uint(tmp % 100)
	}

	return version, nil
}

func (version VersionNumber) pack() uint32 {
	var packed uint32 = 0
	for _, section := range []uint{version.major, version.minor, version.patch, version.subpatch} {
		packed *= 100
		packed += uint32(section % 100)
	}
	return packed
}

func (version VersionNumber) String() string {
	verstr := fmt.Sprintf("%d.%d.%d", version.major, version.minor, version.patch)
	if version.subpatch != 0 {
		verstr = fmt.Sprintf("%s.%d", verstr, version.subpatch)
	}
	return verstr
}
