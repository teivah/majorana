package collections

func FirstMapElement[A comparable, B any](m map[A]B) (a A, b B, exists bool) {
	if len(m) == 0 {
		return
	}
	for k, v := range m {
		return k, v, true
	}
	return
}
