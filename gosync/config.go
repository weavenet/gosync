package gosync

type config struct {
	Source     string
	Target     string
	Concurrent int
}

func NewConfig() *config {
	return &config{}
}
