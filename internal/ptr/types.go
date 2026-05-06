package ptr

func Deref[T any](v *T) T {
	if v == nil {
		var zero T
		return zero
	}

	return *v
}

// Pointer returns a pointer to v.
// TODO(ldez) it must be replaced with the builtin 'new' function when min Go 1.26.
func Pointer[T any](v T) *T { return &v }
