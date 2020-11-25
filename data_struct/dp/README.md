# 动态规划

* 1.根据题目含义写出状态转换方程，类似数学题目里的递推式
* 2.根据递推式写出dp，通常是对二维数组dp[m][n]进行从上到下从做到右的赋值
* 3.赋值完整个dp数组之后，即可得到最终结果

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