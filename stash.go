package squirrel

import "time"

type Stash struct {
	status StashStatus
	val interface{}
}

type StashStatus struct {
	creation time.Time
}

func NewStash(val interface{}) *Stash {
	return &Stash{
		val: val,
	}
}

func (s *Stash) Now() *Stash {
	s.status.creation = time.Now()
	return s
}

func (s *Stash) CreatedAt(t time.Time) *Stash {
	s.status.creation = t
	return s
}