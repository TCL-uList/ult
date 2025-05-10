package version

import (
	"fmt"
	"strings"
)

type BumpType int

const (
	BumpTypeYear BumpType = iota
	BumpTypeMajor
	BumpTypeMilestone
	BumpTypeMinor
	BumpTypeBuild
)

func (b BumpType) String() string {
	switch b {
	case BumpTypeYear:
		return "year"
	case BumpTypeMilestone:
		return "milestone"
	case BumpTypeMajor:
		return "major"
	case BumpTypeMinor:
		return "minor"
	case BumpTypeBuild:
		return "build"
	}

	msg := fmt.Sprintf("invalid build type: %T", b)
	panic(msg)
}

func ParseBumpType(s string) (BumpType, error) {
	switch strings.ToLower(s) {
	case "year":
		return BumpTypeYear, nil
	case "milestone":
		return BumpTypeMilestone, nil
	case "major":
		return BumpTypeMajor, nil
	case "minor":
		return BumpTypeMinor, nil
	case "build":
		return BumpTypeBuild, nil
	}

	return BumpTypeYear, fmt.Errorf("invalid bump type: %s", s)
}
