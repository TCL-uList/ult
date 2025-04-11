package assignee

import "fmt"

type Assignee struct {
	Name  string
	Email string
}

func (a Assignee) String() string {
	return fmt.Sprintf("Name: %s, Email: %s", a.Name, a.Email)
}
