package data

import "slices"

type Assignees struct {
	Nodes []Assignee
}

type Assignee struct {
	Login string
}

// AssigneesFromLogins creates an Assignees struct from a slice of login names.
func AssigneesFromLogins(logins []string) Assignees {
	nodes := make([]Assignee, len(logins))
	for i, login := range logins {
		nodes[i] = Assignee{Login: login}
	}
	return Assignees{Nodes: nodes}
}

// AddAssignees returns a new slice with addedAssignees appended (deduped).
func AddAssignees(assignees, addedAssignees []Assignee) []Assignee {
	result := assignees
	for _, a := range addedAssignees {
		if !slices.Contains(result, a) {
			result = append(result, a)
		}
	}
	return result
}

// RemoveAssignees returns a new slice with removedAssignees filtered out.
func RemoveAssignees(assignees, removedAssignees []Assignee) []Assignee {
	result := make([]Assignee, 0, len(assignees))
	for _, a := range assignees {
		if !slices.Contains(removedAssignees, a) {
			result = append(result, a)
		}
	}
	return result
}
