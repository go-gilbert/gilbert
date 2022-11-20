package options

// Option is a function option
type Option[T any] func(opt *T)

// Apply runs a sequence of option functions over a passed value
func Apply[T any](target *T, opts []Option[T]) {
	for _, opt := range opts {
		opt(target)
	}
}
