# 动态规划

* 1.根据题目含义写出dp含义，状态转换方程，类似数学题目里的递推式，有的题目可能有多个dp式子，如题四
* 2.根据递推式写出dp，通常是对二维数组dp[m][n]进行从上到下从左到右的赋值,简单的题目用一维数组就够了
* 3.赋值完整个dp数组之后，即可得到最终结果
* 4.多做几个dp题目就可以发现一些规律了
* 1. 需要让dp[i][j]与dp[i-1][j-1]建立联系，得保证一些边界条件，像序列以i,j结尾这样
  2. 最终想得到的结果可以在dp循环赋值时一遍迭代一遍比较，等dp赋值结束，结果也得到了

## 题目一：编辑距离[leetcode-72](https://leetcode-cn.com/problems/edit-distance/)

### 问题描述

给定两个单词 word1 和 word2，计算出将 word1 转换成 word2 所使用的最少操作数 。

你可以对一个单词进行如下三种操作：

* 插入一个字符 
* 删除一个字符 
* 替换一个字符


    示例：
    输入: word1 = "horse", word2 = "ros"
    输出: 3
    解释: 
    horse -> rorse (将 'h' 替换为 'r')
    rorse -> rose (删除 'r')
    rose -> ros (删除 'e')
    
### 定义dp数组含义
可以从字符串的最左侧开始，word1[:i]转换成word2[:j]最少需要多少步,值为dp[i][j]
用上面的word1 = "horse", word2 = "ros"举例
dp[0][0]=0
dp[1][0]=1 // 删掉h 'h'->''
dp[0][1]=1 // 插入r ''->'r'
dp[1][1]=1

### 写出状态转换方程

* 当word1[i]=word2[j]时，不用做任何调整,dp[i][j]=dp[i-1][j-1]
* 当word1[i]!=word2[j]时，需要调整保证word1[i]=word2[j]
* 1. word1[:i]末尾插入一个字符word2[j],需要先将word1[:i]转换成word2[:j-1], dp[i][j]= dp[i][j-1]+1
  2. 把word1[i]替换成word2[j],需要先将word1[:i-1]转换成word2[:j-1], dp[i][j]=dp[i-1][j-1]+1
  3. 删除word1[i]，需要先将word[:i-1]转换成word2[:j], dp[i][j]=dp[i-1][j]+1

所以，
    
    if word1[i]==word2[j]
        dp[i][j]=dp[i-1][j-1]
    else 
        dp[i][j]=min{dp[i][j-1]+1,dp[i-1][j-1]+1,dp[i-1][j]+1}

就这样完成了状态方程，可以直接撸代码了
### code

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

## 题目二：最长重复子数组[leetcode](https://leetcode-cn.com/problems/maximum-length-of-repeated-subarray/)
### 问题描述

给两个整数数组 A 和 B ，返回两个数组中公共的、长度最长的子数组的长度。

示例：

    输入：
    A: [1,2,3,2,1]
    B: [3,2,1,4,7]
    输出：3
    解释：
    长度最长的公共子数组是 [3, 2, 1] 。
 

提示：

* 1 <= len(A), len(B) <= 1000
* 0 <= A[i], B[i] < 100

### 定义dp数组含义

可以从左侧开始，两个数组各截取1段, A[:i]和B[:j]最长的公共子数组叫dp[i][j]，并且满足条件公共子数组在A,B下标为i,j,这样
可以保证dp[i][j]和dp[i-1][j-1]可以产生联系，便于输出状态转换方程

* i=1,j=2时,[1],[3,2]对应的公共子数组为空数组[]
* i=3,j=1时,[1,2,3],[3]对应的公共子数组为[3]

最终的最长公共子数组长度为max{dp[0...i][0...j]}

### 状态转换方程
* 初始条件dp[0][j]=0,dp[i][0]=0
* 如果A[i]!=B[j],不可能存在以i,j结尾的公共子数组，所以dp[i][j]=0
* 如果A[i]=B[j], dp[i][j]=dp[i-1][j-1]+1

### code

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


## 题目三：最长上升子序列[leetcode-300](https://leetcode-cn.com/problems/longest-increasing-subsequence/)

### 问题描述

给定一个无序的整数数组，找到其中最长上升子序列的长度。

示例:

    输入: [10,9,2,5,3,7,101,18]
    输出: 4 
    解释: 最长的上升子序列是 [2,3,7,101]，它的长度是 4。
    
说明:

* 可能会有多种最长上升子序列的组合，你只需要输出对应的长度即可。
* 你算法的时间复杂度应该为 O(n2) 。

进阶: 你能将算法的时间复杂度降低到 O(n log n) 吗?

### 定义dp数组含义
很容易想到使用dp[i]表示以a[i]结尾的最长上升子数组的长度，
即dp[i]表示a[1...i]数组中最长升序的数组，最后一个元素是a[i]

### 状态转换方程

dp[i]=1+max{dp[i-k]}(a[i]>a[i-k], k from 1 to i)

### code

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

 
进阶的方法要将时间复杂度优化到O(nlogn),显然需要用到二分查找，在此不赘述。

## 题目四：两个子序列的最大点积[leetcode-1458](https://leetcode-cn.com/problems/max-dot-product-of-two-subsequences/)
### 问题描述
给你两个数组 nums1 和 nums2 。

请你返回 nums1 和 nums2 中两个长度相同的 非空 子序列的最大点积。

数组的非空子序列是通过删除原数组中某些元素（可能一个也不删除）后剩余数字组成的序列，
但不能改变数字间相对顺序。比方说，[2,3,5] 是 [1,2,3,4,5] 的一个子序列而 [1,5,3] 不是。

示例 1：

    输入：nums1 = [2,1,-2,5], nums2 = [3,0,-6]
    输出：18
    解释：从 nums1 中得到子序列 [2,-2] ，从 nums2 中得到子序列 [3,-6] 。
    它们的点积为 (2*3 + (-2)*(-6)) = 18 。

示例 2：

    输入：nums1 = [3,-2], nums2 = [2,-6,7]
    输出：21
    解释：从 nums1 中得到子序列 [3] ，从 nums2 中得到子序列 [7] 。
    它们的点积为 (3*7) = 21 。
    
示例 3：

    输入：nums1 = [-1,-1], nums2 = [1,1]
    输出：-1
    解释：从 nums1 中得到子序列 [-1] ，从 nums2 中得到子序列 [1] 。
    它们的点积为 -1 。
 

提示：

    1 <= nums1.length, nums2.length <= 500
    -1000 <= nums1[i], nums2[i] <= 100
 

点积：

    定义 a = [a1, a2,…, an] 和 b = [b1, b2,…, bn] 的点积为：

    a.b=a1b1+a2b2+...+anbn

### 定义dp数组含义
* dp[i][j]表示 [a1,a2,...,ai] 和[b1,b2,...,bj]两个数组的非空子序列(以ai,bj结尾)的最大点积，为了让dp[i-1][j-1]和dp[i][j]
建立关系，还得要求子序列以ai,bj结尾
* maxRangeDP[i][j]表示[a1,a2,...,ai] 和[b1,b2,...,bj]两个数组的子序列(以ai,bj结尾)的最大点积，空序列点击为0

### 状态转换方程
* 由于子序列要求非空，则dp[1][1]=a1*b1,而不是0
* 下标从2开始计算dp，以a[i],b[j]结尾的子序列要想点积最大，则要求前面的序列点击最大，即 dp[i][j] = ai*bj + maxRangeDP[i-1][j-1]
* maxRangeDP[i][j] = max{maxRangeDP[i-1][j-1], maxRangeDP[i][j-1], maxRangeDP[i-1][j], dp[i][j]}

### code

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
