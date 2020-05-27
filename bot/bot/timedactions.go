package bot

type TimedAction struct {
	Name       string
	Type       string
	Cooldown   int64
	Messages   []string
	LastCalled int64
}
