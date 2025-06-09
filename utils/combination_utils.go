package utils

func Binom(n, k int) int {
	if k < 0 || k > n {
		return 0
	}
	if k == 0 || k == n {
		return 1
	}
	k = min(k, n-k)
	result := 1
	for i := 0; i < k; i++ {
		result *= (n - i)
		result /= (i + 1)
	}
	return result
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func GenerateCombinations(ids []int, k int) [][]int {
	var result [][]int
	if k == 0 {
		return [][]int{{}}
	}
	var current []int
	var backtrack func(start int)

	backtrack = func(start int) {
		if len(current) == k {
			combination := make([]int, k)
			copy(combination, current)
			result = append(result, combination)
			return
		}
		for i := start; i < len(ids); i++ {
			current = append(current, ids[i])
			backtrack(i + 1)
			current = current[:len(current)-1]
		}
	}

	backtrack(0)
	return result
}
