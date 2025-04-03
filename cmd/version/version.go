package version_command

import "fmt"

// Version represents semantic versioning with build number
type Version struct {
	Major int
	Minor int
	Patch int
	Build int
}

// String returns formatted version string (implements Stringer interface)
func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%02d+%02d", v.Major, v.Minor, v.Patch, v.Build)
}
