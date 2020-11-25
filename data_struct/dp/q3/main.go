package main

import "fmt"

func main() {
	fmt.Println(lengthOfLIS([]int{10, 9, 2, 5, 3, 7, 101, 18}))
}

func lengthOfLIS(nums []int) int {
	l := len(nums)
	// 长度=0时结果为0，长度为1时，结果为1
	if l <= 1 {
		return l
	}
	// 初始化dp数组
	dp := make([]int, l+1)
	max := 1
	for i := 1; i <= l; i++ {
		dp[i] = 1 // 算上自身
		maxL := 0
		// 从右往左遍历，查找元素小于nums[i]的最长子序列
		for k := i - 1; k >= 1; k-- {
			if nums[i-1] > nums[k-1] && maxL < dp[k] {
				maxL = dp[k]
			}
		}
		dp[i] += maxL

		// 比较dp[i]的最大值
		if max < dp[i] {
			max = dp[i]
		}
	}
	return max
}
