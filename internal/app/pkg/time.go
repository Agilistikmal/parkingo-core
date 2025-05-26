package pkg

import (
	"strings"
	"time"
)

func GetCurrentTime() time.Time {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	return time.Now().In(loc)
}

func CalculateSimilarity(a, b string) float64 {
	a = strings.ToLower(a)
	b = strings.ToLower(b)

	lenA := len(a)
	lenB := len(b)

	if lenA == 0 || lenB == 0 {
		return 0
	}

	matrix := make([][]int, lenA+1)
	for i := range matrix {
		matrix[i] = make([]int, lenB+1)
	}

	for i := 0; i <= lenA; i++ {
		matrix[i][0] = i
	}

	for j := 0; j <= lenB; j++ {
		matrix[0][j] = j
	}

	for i := 1; i <= lenA; i++ {
		for j := 1; j <= lenB; j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}

			matrix[i][j] = min(matrix[i-1][j]+1, matrix[i][j-1]+1, matrix[i-1][j-1]+cost)
		}
	}

	distance := matrix[lenA][lenB]
	similarity := 1 - float64(distance)/float64(max(lenA, lenB))
	return similarity
}
