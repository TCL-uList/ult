package version

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

// Version represents semantic versioning with build number
type Version struct {
	Major int
	Minor int
	Patch int
	Build int
}

// String returns formatted version string like "2010.200.01+04"
func (v Version) String() string {
	return fmt.Sprintf("%04d.%03d.%02d+%02d", v.Major, v.Minor, v.Patch, v.Build)
}

// String returns formatted version string without the build part like "2010.200.01"
func (v Version) StringNoBuild() string {
	return fmt.Sprintf("%04d.%03d.%02d", v.Major, v.Minor, v.Patch)
}

// Parse parses a version line like "version: 10.20.03+04" into a Version struct.
// It expects the version to be in the format "major.minor.patch+build" where all
// components are integers. Returns an error if the format is invalid or any component
// cannot be parsed as an integer.
func Parse(line string) (*Version, error) {
	if len(line) == 0 {
		return nil, errors.New("Version line string cannot be empty")
	}
	re := regexp.MustCompile(`(\d+)\.(\d+)\.(\d+)\+(\d+)`)
	matches := re.FindStringSubmatch(line)
	if matches == nil || len(matches) != 5 {
		return nil, fmt.Errorf("Version string doesn't match expected format \"version: 2020.100.01+01\", got: %s", line)
	}

	major, err := strconv.Atoi(matches[1])
	if err != nil {
		return nil, fmt.Errorf("invalid major version (%s): %q", matches[1], err)
	}

	minor, err := strconv.Atoi(matches[2])
	if err != nil {
		return nil, fmt.Errorf("invalid minor version (%s): %q", matches[2], err)
	}

	patch, err := strconv.Atoi(matches[3])
	if err != nil {
		return nil, fmt.Errorf("invalid patch version (%s): %q", matches[3], err)
	}

	build, err := strconv.Atoi(matches[4])
	if err != nil {
		return nil, fmt.Errorf("invalid build number (%s): %q", matches[4], err)
	}

	version := Version{
		Major: major,
		Minor: minor,
		Patch: patch,
		Build: build,
	}

	return &version, nil
}

// Bump updates the provided Version struct based on the bump type specified
// It supports bumping the build number, patch, minor, or major version components.
func (v *Version) Bump(bumpType BumpType) {
	switch bumpType {
	case BumpTypeBuild:
		v.Build++

	case BumpTypePatch:
		v.Build = 1
		v.Patch++

	case BumpTypeMinor:
		v.Build = 1
		v.Patch = 1
		v.Minor++

	case BumpTypeMajor:
		v.Build = 1
		v.Patch = 1
		v.Minor = v.Minor - (v.Minor % 100) + 100
		v.Major++
	}
}

// FetchFromLines searches through the lines to find the version line.
// Will parses it, and returns the Version struct along with the line index where it was found.
// Returns an error if the version line is not found or if parsing fails.
func FetchFromLines(lines []string) (*Version, int, error) {
	for i, line := range lines {
		// the minimum string containing a version (0.0.0+00) has 8 characters
		if len(line) < 8 {
			continue
		}

		version, err := Parse(line)
		if err != nil {
			continue
		}

		return version, i, nil
	}

	return nil, -1, errors.New("version not found in pubspec.yaml")
}
