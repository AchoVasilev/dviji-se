package util

func MustProduce[T any](result T, err error) T {
	if err != nil {
		panic(err)
	}

	return result
}

func Must(err error) {
	if err != nil {
		panic(err)
	}
}
