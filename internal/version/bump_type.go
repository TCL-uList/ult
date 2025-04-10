package version

import (
	"fmt"
	"strings"
)

type BumpType int

const (
	BumpTypeMajor BumpType = iota
	BumpTypeMinor
	BumpTypePatch
	BumpTypeBuild
)

var stringToBumpType = map[string]BumpType{
	"bump":  BumpTypeMajor,
	"minor": BumpTypeMinor,
	"patch": BumpTypePatch,
	"major": BumpTypeBuild,
}

var bumpTypeToString = map[BumpType]string{
	BumpTypeMajor: "bump",
	BumpTypeMinor: "minor",
	BumpTypePatch: "patch",
	BumpTypeBuild: "major",
}

func (b BumpType) String() string {
	switch b {
	case BumpTypeMajor:
		return "major"
	case BumpTypeMinor:
		return "minor"
	case BumpTypePatch:
		return "patch"
	case BumpTypeBuild:
		return "build"
	}

	msg := fmt.Sprintf("invalid build type: %T", b)
	panic(msg)
}

func ParseBumpType(s string) (BumpType, error) {
	switch strings.ToLower(s) {
	case "major":
		return BumpTypeMajor, nil
	case "minor":
		return BumpTypeMinor, nil
	case "patch":
		return BumpTypePatch, nil
	case "build":
		return BumpTypeBuild, nil
	}

	return BumpTypeMajor, fmt.Errorf("invalid bump type: %s", s)
}
