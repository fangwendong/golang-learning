package main

import "fmt"

func main() {
	fmt.Println(maxDotProduct([]int{2, 1, -2, 5}, []int{3, 0, -6}))
	fmt.Println(maxDotProduct([]int{3, -2}, []int{2, -6, 7}))
	fmt.Println(maxDotProduct([]int{-1, -1}, []int{1, 1}))
	fmt.Println(maxDotProduct([]int{5, -4, -3}, []int{-4, -3, 0, -4, 2}))
}

func maxDotProduct(nums1 []int, nums2 []int) int {
	l1 := len(nums1)
	l2 := len(nums2)

	// 构建dp[l1][l2]
	dp := make([][]int, l1+1)
	maxRangeDP := make([][]int, l1+1) // 表示a[1..i]到b[1..j]子序列点击最大值，空序列点击为0
	for i := 0; i <= l1; i++ {
		dp[i] = make([]int, l2+1)
		maxRangeDP[i] = make([]int, l2+1)
	}

	max := nums1[0] * nums2[0] // 表示a[]和b[]中非空子序列最大点积
	// 根据状态转换方程给dp赋值
	for i := 1; i <= l1; i++ {
		for j := 1; j <= l2; j++ {
			maxLeftDot := maxRangeDP[i-1][j-1] // maxLeftDot表示a[1..i],b[1...j]中子序列最大点积,可以为空(maxLeftDot=0)
			dp[i][j] = nums1[i-1] * nums2[j-1] // 以nums1[i]和nums2[j]结尾的子序列，必须带上他俩的点积
			dp[i][j] += maxLeftDot
			maxRangeDP[i][j] = maxF(maxRangeDP[i-1][j-1], maxRangeDP[i][j-1], maxRangeDP[i-1][j], dp[i][j])
			max = maxF(max, dp[i][j])
		}
	}

	return max
}

func maxF(args ...int) int {
	m := args[0]
	for i := 1; i < len(args); i++ {
		if args[i] > m {
			m = args[i]
		}
	}
	return m
}
