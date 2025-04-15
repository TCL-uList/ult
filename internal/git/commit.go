package git

import (
	"fmt"
	"time"

	"ulist.app/ult/internal/assignee"
)

type Commit struct {
	Assignee assignee.Assignee
	Hash     string
	Date     time.Time
	Message  string
}

func (c Commit) String() string {
	return fmt.Sprintf(
		"Hash: %s\nAuthor: %s <%s>\nDate: %s\nMessage: %s",
		c.Hash,
		c.Assignee.Name,
		c.Assignee.Email,
		c.Date,
		c.Message,
	)
}
