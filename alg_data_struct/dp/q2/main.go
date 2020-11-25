package main

import "fmt"

func main() {
	fmt.Println(findLength([]int{1, 2, 3, 2, 1}, []int{3, 2, 1, 4, 7}))
	fmt.Println(findLength([]int{1}, []int{3}))
}

func findLength(A []int, B []int) int {
	l1 := len(A)
	l2 := len(B)
	// 初始化dp数组
	dp := make([][]int, l1+1)
	for i := 0; i <= l1; i++ {
		dp[i] = make([]int, l2+1)
	}

	// 从上到下，从左到右赋值,maxLength记录最大的公共子数组长度
	var maxLength int
	for i := 1; i <= l1; i++ {
		for j := 1; j <= l2; j++ {
			if A[i-1] == B[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
				if dp[i][j] > maxLength {
					maxLength = dp[i][j]
				}
			}
		}
	}
	return maxLength
}
