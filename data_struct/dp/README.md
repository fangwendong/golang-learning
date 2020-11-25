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

