package stage

//go:generate stringer -type=Stage

type Stage int

const (
	Input Stage = iota
	Check
	Files
	Create
	Upload
	Delete
)
