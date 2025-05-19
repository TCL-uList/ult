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
				Year:  2000,
				Major: 200,
				Minor: 2,
				Build: 3,
			},
		},
		{
			name:  "bumping minor",
			input: BumpTypeMinor,
			want: Version{
				Year:  2000,
				Major: 200,
				Minor: 3,
				Build: 1,
			},
		},
		{
			name:  "bumping milestone",
			input: BumpTypeMilestone,
			want: Version{
				Year:  2000,
				Major: 300,
				Minor: 1,
				Build: 1,
			},
		},
		{
			name:  "bumping major",
			input: BumpTypeMajor,
			want: Version{
				Year:  2000,
				Major: 201,
				Minor: 1,
				Build: 1,
			},
		},
		{
			name:  "bumping year",
			input: BumpTypeYear,
			want: Version{
				Year:  2001,
				Major: 100,
				Minor: 1,
				Build: 1,
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			version := Version{Year: 2000, Major: 200, Minor: 2, Build: 2}
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
			input: "1000.200.30+456",
			want: Version{
				Year:  1000,
				Major: 200,
				Minor: 30,
				Build: 456,
			},
		},
		{
			name:    "invalid format",
			input:   "1.2",
			wantErr: true,
		},
		{
			name:  "invalid semver",
			input: "1000.200.30+456",
			want: Version{
				Year:  1000,
				Major: 200,
				Minor: 30,
				Build: 456,
			},
		},
		{
			name:  "with metadata",
			input: "1000.200.30+456",
			want: Version{
				Year:  1000,
				Major: 200,
				Minor: 30,
				Build: 456,
			},
		},
		{
			name:  "valid semver with prefix",
			input: "version: 1000.200.30+456",
			want: Version{
				Year:  1000,
				Major: 200,
				Minor: 30,
				Build: 456,
			},
		},
		{
			name:  "valid semver with postfix",
			input: "1000.200.30+456 postfix",
			want: Version{
				Year:  1000,
				Major: 200,
				Minor: 30,
				Build: 456,
			},
		},
		{
			name:  "valid semver with actual value",
			input: "version: 2025.200.01+10",
			want: Version{
				Year:  2025,
				Major: 200,
				Minor: 01,
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

func TestFetchFromLines(t *testing.T) {
	tt := []struct {
		name    string
		input   []string
		want    *Version
		wantIdx int
		wantErr bool
	}{
		{
			name: "no version in list",
			input: []string{
				"aa",
				"bb",
				"cc",
			},
			wantErr: true,
		},
		{
			name: "no version in list with empty space",
			input: []string{
				"",
				"----",
				"\n",
			},
			wantErr: true,
		},
		{
			name: "version at end",
			input: []string{
				"",
				"----",
				"\n",
				"version: 2000.200.02+02",
			},
			want:    &Version{Year: 2000, Major: 200, Minor: 02, Build: 02},
			wantIdx: 3,
			wantErr: false,
		},
		{
			name: "version at end with no prefix",
			input: []string{
				"",
				"----",
				"2000.200.02+02",
				"\n",
			},
			want:    &Version{Year: 2000, Major: 200, Minor: 02, Build: 02},
			wantIdx: 2,
			wantErr: false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			got, idx, err := FetchFromLines(tc.input)
			if (err != nil) != tc.wantErr {
				t.Fatalf("ParseVersion() error = %v, wantErr %v", err, tc.wantErr)
			}
			if !tc.wantErr && !versionsEqual(*got, *tc.want) {
				t.Errorf("ParseVersion() = %+v, want %+v", got, tc.want)
			}
			if !tc.wantErr && idx != tc.wantIdx {
				t.Errorf("FetchFromLines() idx = %d, want %+v", idx, tc.wantIdx)
			}
		})
	}
}

func versionsEqual(a, b Version) bool {
	return a.Year == b.Year &&
		a.Major == b.Major &&
		a.Minor == b.Minor &&
		a.Build == b.Build
}
