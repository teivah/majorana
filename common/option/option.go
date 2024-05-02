package option

type Optional[T any] struct {
	value  T
	exists bool
}

func Of[T any](t T) Optional[T] {
	return Optional[T]{
		value:  t,
		exists: true,
	}
}

func None[T any]() Optional[T] {
	return Optional[T]{}
}

func (o Optional[T]) Get() (T, bool) {
	return o.value, o.exists
}
