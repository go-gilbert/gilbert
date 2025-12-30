package argschema

type Option[T any] struct {
	ok  bool
	val T
}

func (opt Option[T]) Value() (T, bool) {
	return opt.val, opt.ok
}

func Some[T any](v T) Option[T] {
	return Option[T]{
		ok:  true,
		val: v,
	}
}

func None[T any]() Option[T] {
	return Option[T]{}
}
