package git

import (
	"testing"
	"time"
)

func TestGetCommitInfoFromStdout_StandartCommit(t *testing.T) {
	input := `commit 3a4f5d6e7g8h9i0j1k2l3m4n5o6p7q8r9s0
Author: John Doe <john@example.com>
Date:   Wed Apr 10 15:30:45 2024 -0300

	Add new feature implementation
`

	got, err := getCommitInfoFromStdout(input)
	if err != nil {
		t.Errorf("getCommitInfoFromStdout() error = %v", err)
		return
	}

	expected := &Commit{
		Hash:        "3a4f5d6e7g8h9i0j1k2l3m4n5o6p7q8r9s0",
		AuthorName:  "John Doe",
		AuthorEmail: "john@example.com",
		Date:        time.Date(2024, time.April, 10, 15, 30, 45, 0, time.FixedZone("", -3*60*60)),
		Message:     "Add new feature implementation",
	}

	assertCommitEqual(t, got, expected)
}

func TestGetCommitInfoFromStdout_MultiLineMessage(t *testing.T) {
	input := `commit a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6
Author: Jane Smith <jane@company.com>
Date:   Tue Mar 5 09:15:30 2024 +0000

	Refactor database module
	
	- Improved connection handling
	- Added transaction support
	- Fixed memory leaks
`

	got, err := getCommitInfoFromStdout(input)
	if err != nil {
		t.Errorf("getCommitInfoFromStdout() error = %v", err)
		return
	}

	expected := &Commit{
		Hash:        "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6",
		AuthorName:  "Jane Smith",
		AuthorEmail: "jane@company.com",
		Date:        time.Date(2024, time.March, 5, 9, 15, 30, 0, time.UTC),
		Message:     "Refactor database module",
	}

	assertCommitEqual(t, got, expected)
}

func TestGetCommitInfoFromStdout_MinimalCommit(t *testing.T) {
	input := `commit b5c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7
Author: CI Bot <ci@example.com>
Date:   Mon Jan 1 00:00:00 2024 +0000

`
	got, err := getCommitInfoFromStdout(input)
	if err != nil {
		t.Errorf("getCommitInfoFromStdout() error = %v", err)
		return
	}

	expected := &Commit{
		Hash:        "b5c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7",
		AuthorName:  "CI Bot",
		AuthorEmail: "ci@example.com",
		Date:        time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
		Message:     "",
	}

	assertCommitEqual(t, got, expected)
}

func TestGetCommitInfoFromStdout_InvalidCommitFormat(t *testing.T) {
	input := "not a valid commit message"

	_, err := getCommitInfoFromStdout(input)
	if err == nil {
		t.Error("was expecting a error due to invalid commit format")
		return
	}
}

func assertCommitEqual(t *testing.T, actual, expected *Commit) {
	t.Helper()

	if actual.Hash != expected.Hash {
		t.Errorf("Hash mismatch: got %q, want %q", actual.Hash, expected.Hash)
	}
	if actual.AuthorName != expected.AuthorName {
		t.Errorf("AuthorName mismatch: got %q, want %q", actual.AuthorName, expected.AuthorName)
	}
	if actual.AuthorEmail != expected.AuthorEmail {
		t.Errorf("AuthorEmail mismatch: got %q, want %q", actual.AuthorEmail, expected.AuthorEmail)
	}
	if !actual.Date.Equal(expected.Date) {
		t.Errorf("Date mismatch: got %v, want %v", actual.Date, expected.Date)
	}
	if actual.Message != expected.Message {
		t.Errorf("Message mismatch:\ngot:\n%q\nwant:\n%q", actual.Message, expected.Message)
	}
}
