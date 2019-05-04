package pkga

// A HelloCallback demonstrates perhaps the coolest gomobile feature
type HelloCallback interface {
	// YourName returns a string to say hello to
	YourName() string
}

// NiceCallback takes a callback and invokes it
func NiceCallback(cb HelloCallback) string {
	return "hello " + cb.YourName()
}
