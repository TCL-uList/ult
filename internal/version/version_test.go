package version

import (
	"testing"
)

func TestBump(t *testing.T) {
	tt := []struct {
		name  string
		input BumpType
		want  Version
	}{
		{
			name:  "bumping build",
			input: BumpTypeBuild,
			want: Version{
				Major: 2000,
				Minor: 200,
				Patch: 2,
				Build: 3,
			},
		},
		{
			name:  "bumping patch",
			input: BumpTypePatch,
			want: Version{
				Major: 2000,
				Minor: 200,
				Patch: 3,
				Build: 1,
			},
		},
		{
			name:  "bumping minor",
			input: BumpTypeMinor,
			want: Version{
				Major: 2000,
				Minor: 201,
				Patch: 1,
				Build: 1,
			},
		},
		{
			name:  "bumping major",
			input: BumpTypeMajor,
			want: Version{
				Major: 2001,
				Minor: 300,
				Patch: 1,
				Build: 1,
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			version := Version{Major: 2000, Minor: 200, Patch: 2, Build: 2}
			version.Bump(tc.input)
			if !versionsEqual(version, tc.want) {
				t.Errorf("ParseVersion() = %+v, want %+v", version, tc.want)
			}
		})
	}
}

func TestParse(t *testing.T) {
	tt := []struct {
		name    string
		input   string
		want    Version
		wantErr bool
	}{
		{
			name:  "valid semver",
			input: "1.2.3+456",
			want: Version{
				Major: 1,
				Minor: 2,
				Patch: 3,
				Build: 456,
			},
		},
		{
			name:    "invalidjformat",
			input:   "1.2",
			wantErr: true,
		},
		{
			name:  "with metadata",
			input: "1.2.3+456",
			want: Version{
				Major: 1,
				Minor: 2,
				Patch: 3,
				Build: 456,
			},
		},
		{
			name:  "valid semver with prefix",
			input: "version: 1.2.3+456",
			want: Version{
				Major: 1,
				Minor: 2,
				Patch: 3,
				Build: 456,
			},
		},
		{
			name:  "valid semver with postfix",
			input: "1.2.3+456 postfix",
			want: Version{
				Major: 1,
				Minor: 2,
				Patch: 3,
				Build: 456,
			},
		},
		{
			name:  "valid semver with actual value",
			input: "version: 2025.200.01+10",
			want: Version{
				Major: 2025,
				Minor: 200,
				Patch: 01,
				Build: 10,
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Parse(tc.input)
			if (err != nil) != tc.wantErr {
				t.Fatalf("ParseVersion() error = %v, wantErr %v", err, tc.wantErr)
			}
			if !tc.wantErr && !versionsEqual(*got, tc.want) {
				t.Errorf("ParseVersion() = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func versionsEqual(a, b Version) bool {
	return a.Major == b.Major &&
		a.Minor == b.Minor &&
		a.Patch == b.Patch &&
		a.Build == b.Build
}
