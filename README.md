# 带盐的慢哈希值示例

## 名词解释

### 哈希算法
哈希算法(Hash Algorithm)是一种将输入数据(也称为消息或文本)转换为固定长度的字符串(通常是固定位数的二进制或十六进制数字)的算法。哈希算法的主要目的是通过一种高效的方式，将任意长度的输入数据映射到固定长度的输出，通常称为哈希值。哈希值的特点是：
1. 相同的输入将始终产生相同的哈希值。
2. 即使输入数据的微小变化，哈希值也会显著变化。
3. 不容易从哈希值反向推导出原始输入数据。

### 加盐哈希
加盐哈希就是在计算密码的哈希值时，在密码字符串前/后面添加一个称为“盐(salt)”的随机字符串，这个随机字符串称为盐值，它的作用是增加哈希后密码的随机性

### 快哈希算法:
快哈希算法(Fast Hash Algorithm)是指计算速度相对较快的哈希算法。这些算法通常设计用于快速计算哈希值，适用于对性能要求较高的应用。常见的快哈希算法包括MD5、SHA-1、SHA-256 等。然而，这些算法在安全性方面存在问题，因为它们容易受到碰撞攻击(两个不同的输入产生相同的哈希值)的影响，因此在安全敏感的场景中不再推荐使用。
常见的快哈希算法有MD5, SHA256等

### 慢哈希算法
慢哈希算法(Slow Hash Algorithm)是指计算速度相对较慢的哈希算法，通常被用于存储密码和保护用户凭据。这些算法故意设计成计算速度较慢，以增加密码破解的难度，提高密码的安全性。常见的慢哈希算法包括Bcrypt、Scrypt、Argon2、PBKDF2 等。
特征如下:
1. 计算速度更慢，需要消耗更多CPU和内存资源，从而对抗硬件加速攻击；
2. 使用更复杂的算法，组合密码学原语，增加破解难度；
3. 可以配置资源消耗参数，调整安全强度；
4. 特定优化使并行计算困难；
5. 经过长时间的密码学分析，仍然安全可靠

#### 带盐慢哈希算法

## 测试
慢哈希算法更适合密码哈希的原因是可以大幅增加攻击者密码破解的成本

以SHA256和Scrypt两个算法为例做的一个简单的benchmark来测试进行量化对比
```go
package main
  
import (
    "crypto/sha256"
    "testing"

    "golang.org/x/crypto/scrypt"
)

func BenchmarkSHA256(b *testing.B) {
    b.ReportAllocs()
    data := []byte("hello world")
    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        sha256.Sum256(data)
    }
}

func BenchmarkScrypt(b *testing.B) {
    b.ReportAllocs()
    const keyLen = 32
    data := []byte("hello world")
    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        scrypt.Key(data, data, 16384, 8, 1, keyLen)
    }
}
```

输出的benchmark结果：

```$go test -bench .

BenchmarkSHA256-8     6097324        195.3 ns/op        0 B/op        0 allocs/op
BenchmarkScrypt-8          26   41812138 ns/op 16781836 B/op       22 allocs/op
PASS
ok   demo 2.533s
```
cpu消耗和内存开销，Scrypt算法都是SHA256的几个数量级的倍数

加盐的慢哈希也是目前的主流的用户密码存储方案，那有读者会问：这四个算法选择哪个更佳呢？说实话要想对这个四个算法做个全面的对比，需要很强的密码学专业知识，这里直接给结论(当然也是来自网络资料)：建议使用Scrypt或Argon2系列的算法，它们俩可提供更高的抗ASIC和并行计算能力，Bcrypt由于简单高效和成熟，目前也仍十分流行。

不过，慢哈希算法在给攻击者带来时间和资源成本等困难的同时，也给服务端正常的身份认证带来一定的性能开销，不过大多数开发者认为这种设计取舍是值得的。

下面我们就基于慢哈希算法结合加盐，用实例说明一下一个Web应用的用户注册与登录过程中，密码是如何被存储和用来验证用户身份的

## 使用
本例使用Go官方维护的golang.org/x/crypto为我们提供了高质量的scrypt包，当然crypto下也有bcrypt、argon2和pbkdf2的实现, 

## 说明
全篇参考了和引用[通过实例理解Web应用用户密码存储方案](https://mp.weixin.qq.com/s/_iT9GAtn0hQXDQYW_PbPCA)进行了逻辑代码的修改, 加入了`gorm`
