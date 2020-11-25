package main

import "fmt"

func main() {
	type testArgs struct {
		word1, word2 string
	}
	tests := []testArgs{
		{"horse", "ros"},
		{"intention", "execution"},
		{"horse", "horse"},
		{"", ""},
	}
	for _, t := range tests {
		fmt.Println(t, minDistance(t.word1, t.word2), minDistance(t.word2, t.word1))
	}
}
func minDistance(word1 string, word2 string) int {
	l1 := len(word1)
	l2 := len(word2)
	// 构建dp[l1][l2]
	dp := make([][]int, l1+1)
	for i := 0; i <= l1; i++ {
		dp[i] = make([]int, l2+1)
	}
	// 初始化边界数据i=0,j=0
	for j := 0; j <= l2; j++ {
		dp[0][j] = j
	}
	for i := 0; i <= l1; i++ {
		dp[i][0] = i
	}

	// 从左到右，从上到下根据状态方程赋值
	for i := 1; i <= l1; i++ {
		for j := 1; j <= l2; j++ {
			if word1[i-1] == word2[j-1] {
				dp[i][j] = dp[i-1][j-1]
			} else {
				dp[i][j] = min(dp[i][j-1]+1, dp[i-1][j-1]+1, dp[i-1][j]+1)
			}
		}
	}
	return dp[l1][l2]
}

func min(args ...int) int {
	m := args[0]
	for i := 1; i < len(args); i++ {
		if args[i] < m {
			m = args[i]
		}
	}
	return m
}
