package utils

import (
	"strings"
	"unicode"
)

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

func mapInvisible(r rune) rune {
	if unicode.IsGraphic(r) {
		return r
	}
	return -1
}

// Remove invisible characters from a string.
func RemoveInvisible(text string) string {
	return strings.Map(mapInvisible, text)
}
