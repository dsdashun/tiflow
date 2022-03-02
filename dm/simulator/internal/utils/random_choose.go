package utils

import (
	"math/rand"
)

//weights: key: string key, value: weight
func RandomChooseKeyByWeights(weights map[string]int) string {
	expandedKeys := []string{}
	for k, weight := range weights {
		for i := 0; i < weight; i++ {
			expandedKeys = append(expandedKeys, k)
		}
	}
	idx := rand.Intn(len(expandedKeys))
	return expandedKeys[idx]
}
