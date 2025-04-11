package git

import (
	"fmt"
	"time"
)

type Commit struct {
	AuthorName  string
	AuthorEmail string
	Hash        string
	Date        time.Time
	Message     string
}

func (c Commit) String() string {
	return fmt.Sprintf(
		"Hash: %s\nAuthor: %s <%s>\nDate: %s\nMessage: %s",
		c.Hash,
		c.AuthorName,
		c.AuthorEmail,
		c.Date,
		c.Message,
	)
}
