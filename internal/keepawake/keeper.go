package keepawake

// Keeper holds an OS-level keep-awake request for the current process.
type Keeper interface {
	Start() error
	Stop() error
	Name() string
}

// NewKeeper returns the best OS keep-awake implementation for this platform.
func NewKeeper() Keeper {
	return platformKeeper()
}
