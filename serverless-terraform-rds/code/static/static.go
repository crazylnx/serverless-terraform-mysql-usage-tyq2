package static

const (
	Apply int = iota
	Delete
)

const (
	ApplyComplete = "apply_complete"
	Error         = "error"
	Outputs       = "outputs"
	ApplyStart    = "apply_start"
	ApplyError    = "apply_errored"
)
