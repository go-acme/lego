package internal

type ChallengeOption[T any] func(T) error

func CondOptions[C any, O ChallengeOption[C]](condition bool, opt ...O) O {
	if !condition {
		// NoOp options
		return func(C) error {
			return nil
		}
	}

	return func(chlg C) error {
		for _, opt := range opt {
			err := opt(chlg)
			if err != nil {
				return err
			}
		}

		return nil
	}
}

func LazyCondOption[C any, O ChallengeOption[C]](condition bool, fn func() O) O {
	if !condition {
		// NoOp options
		return func(C) error {
			return nil
		}
	}

	return fn()
}

func CombineOptions[C any, O ChallengeOption[C]](opts ...O) O {
	return func(chlg C) error {
		for _, opt := range opts {
			err := opt(chlg)
			if err != nil {
				return err
			}
		}

		return nil
	}
}
