package utils

type indexedPair[T any] struct {
	index int
	value T
}

func ParallelMap[I, O any](inputs []I, mapper func(input I) O) []O {
	collector := make(chan indexedPair[O])
	for i, input := range inputs {
		go func(idx int, input I) {
			collector <- indexedPair[O]{
				index: idx,
				value: mapper(input),
			}
		}(i, input)
	}
	results := make([]O, len(inputs))
	for i := 0; i < len(inputs); i++ {
		pair := <-collector
		results[pair.index] = pair.value
	}
	return results
}
