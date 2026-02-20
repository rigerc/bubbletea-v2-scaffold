package projector

import "time"

type StatusType int

const (
	StatusClean StatusType = iota
	StatusDirty
	StatusAhead
	StatusBehind
	StatusDiverged
	StatusNoRemote
)

type GitStatus struct {
	Branch           string
	Remote           string
	Uncommitted      int
	Unpushed         int
	Unpulled         int
	LastCommitMsg    string
	LastCommitTime   time.Time
	LastCommitAuthor string
	Status           StatusType
}

type Project struct {
	Name         string
	Path         string
	Language     string
	Size         int64
	LastModified time.Time
	Git          GitStatus
}

func (s StatusType) String() string {
	switch s {
	case StatusClean:
		return "clean"
	case StatusDirty:
		return "dirty"
	case StatusAhead:
		return "ahead"
	case StatusBehind:
		return "behind"
	case StatusDiverged:
		return "diverged"
	case StatusNoRemote:
		return "no-remote"
	default:
		return "unknown"
	}
}
