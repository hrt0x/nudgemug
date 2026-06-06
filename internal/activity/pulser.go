package activity

// Nudger injects a tiny user-activity signal into the current desktop session.
type Nudger interface {
	Name() string
	Nudge() error
}

// NewNudger returns the best activity nudger for the current platform.
func NewNudger() Nudger {
	if nudger := platformNudger(); nudger != nil {
		return nudger
	}
	return UnsupportedNudger{}
}

type UnsupportedNudger struct{}

func (UnsupportedNudger) Name() string { return "unsupported" }
func (UnsupportedNudger) Nudge() error {
	return ErrUnsupported
}
