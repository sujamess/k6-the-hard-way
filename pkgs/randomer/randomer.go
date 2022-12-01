package randomer

import "math/rand"

func Float(min, max float64) float64 {
	return min + (rand.Float64() * (max - min))
}

func Floats(min, max float64, n int) []float64 {
	res := make([]float64, n)
	for i := range res {
		res[i] = Float(min, max)
	}
	return res
}
