package event

const (
	VeryEarly = 0
	Early     = 1
	Moderate  = 2
	Late      = 3
	VeryLate  = 4
	System    = 5
)

type Event interface {
	Name() string
}
