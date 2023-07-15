# A URL shortener in Go - with Gin & Redis

学习参考自[Let's build a URL shortener in Go - Part I (eddywm.com)](https://www.eddywm.com/lets-build-a-url-shortener-in-go/)

在本文档中，我们将探讨如何用**Go** 编程语言编写 url 缩短器，将使用 **Redis** 作为实现中超快速数据检索的存储机制。

==借助该项目学习Gin 和 Redis 的使用。==

## 一、项目设置

### 一. 1.项目设置

让我们设置项目并安装项目构建过程中所需的所有依赖项。

- 初始化 go 项目，请确保您的系统中安装了 Go **1.11+**。

- 安装项目依赖项。

```shell
$ go get github.com/go-redis/redis/v9

$ go get -u github.com/gin-gonic/gin
```

> 注意：在本教程的后续步骤中，您将需要在计算机上安装 **Redis**。
> 如果您的计算机上尚未安装 Redis，您可以从[此处](https://redis.io/download?ref=eddywm.com)的此链接下载它，并按照有关操作系统的说明进行安装

## 二、存储层

### 二. 1.Store Service设置

首先，我们必须在项目中创建我们的存储文件夹，因此进入项目目录，创建一个名为的子目录并继续创建 2 个空的 Go 文件：和（稍后我们将在其中为存储编写单元测试）`store``store_service.go``store_service_test.go`

```bash
└── store
    ├── store_service.go
    └── store_service_test.go
```



- 我们将首先围绕 Redis 设置struct ，struct 将用作持久化和检索应用程序数据映射的接口。

打开文件并填写下面的代码。`store_service.go`

```go
package store

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

// StorageService是一个结构体，它将Redis客户端作为其成员。
type StorageService struct {
	redisClient *redis.Client
}

var (
	// storeService是一个全局变量，它将在整个应用程序中使用。
	storeService = &StorageService{}
	// ctx是一个上下文，它将在整个应用程序中使用。
	ctx = context.Background()
)

// CacheDuration表示缓存的持续时间。
const CacheDuration = 6 * time.Hour
```



- 定义结构后，我们终于可以初始化存储服务，在本例中为我们的 Redis 客户端。

```go
func InitializeStore() *StorageService {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	// 通过调用Ping()方法来检查Redis是否已经启动。
	pong, err := redisClient.Ping(ctx).Result()
	if err != nil {
		panic(fmt.Sprintf("Error init Redis: %v", err))
	}

	fmt.Printf("\nRedis started successfully: pong message = {%s}", pong)
	storeService.redisClient = redisClient
	return storeService
}
```



### 二. 2.Store API 设计和实施

```go
// SaveUrlMapping方法将短网址和原始网址保存到Redis中。
func SaveUrlMapping(shortUrl string, originalUrl string, userId string) {
	err := storeService.redisClient.Set(ctx, shortUrl, originalUrl, CacheDuration).Err()
	if err != nil {
		panic(fmt.Sprintf("Failed saving key url | Error: %v - shortUrl: %s - originalUrl: %s\n", err, shortUrl, originalUrl))
	}
}

// RetrieveInitialUrl方法从Redis中检索原始网址。
func RetrieveInitialUrl(shortUrl string) string {
	result, err := storeService.redisClient.Get(ctx, shortUrl).Result()
	if err != nil {
		panic(fmt.Sprintf("Failed retrieving key url | Error: %v - shortUrl: %s\n", err, shortUrl))
	}
	return result
}
```



### 二. 3.单元和集成测试

为了保留最佳实践并避免将来出现意外回归，我们将不得不考虑存储层实现的单元和集成测试。

现在让我们开始安装这部分所需的测试工具。

```shell
go get github.com/stretchr/testify
```

在文件中`store_service_test.go `添加下面代码

```go
package store

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const UserId = "e0dba740-fc4b-4977-872c-d360239e6b1a"
var testStoreService = &StorageService{}

// 初始化测试
func init() {
	testStoreService = InitializeStore()
}

func TestStoreInit(t *testing.T) {
	// 使用测试包中的assert.True()函数来断言条件testStoreService.redisClient != nil为真。
	assert.True(t, testStoreService.redisClient != nil)
}

func TestInsertionAndRetrieval(t *testing.T) {
	initialLink := "https://met2.fzu.edu.cn/meol/index.do"
	shortURL := "wpwpwp"

	// 保存短网址和原始网址
	SaveUrlMapping(shortURL, initialLink, UserId)

	// 从Redis中检索原始网址
	retrievedUrl := RetrieveInitialUrl(shortURL)

	// 使用测试包中的assert.Equal()函数来断言条件initialLink == retrievedUrl为真。
	assert.Equal(t, initialLink, retrievedUrl)
}

```



## 三、short url生成器

在上一部分我们能够设置、构建和测试链接缩短器的存储层。在这一部分中，我们将专门研究我们将用于散列和处理初始输入或长 url 的算法，使其成为与之对应的更小、更短的映射。

在选择算法时，我们确实要记住许多目标：

- 最终输入应更短：最多 8 个字符
- 应该易于人类阅读，避免混淆字符混淆，这些字符在大多数字体中通常相似。
- 熵应该相当大，以避免在短链接生成中发生冲突。



### 三. 1.生成器算法

在此实现过程中，我们将使用两种主要方案：哈希函数和二进制到文本编码算法。

创建 2 个文件`shorturl_generator.go``shorturl_generator_test.go``shortener`并将它们放在上面的文件夹下后，我们的项目目录结构应该看起来像下面的树：

```bash
├── go.mod
├── go.sum
├── main.go
├── shortener
│   ├── shorturl_generator.go
│   └── shorturl_generator_test.go
└── store
    ├── store_service.go
    └── store_service_test.go
```



### 三. 2.缩短器实现

我们选择  **sha256 **和 **Base58** 来实现算法

> 使用 **Base58** 的理由

==**Base58** 减少了字符输出中的混乱==

- 字符 **0，O、I、l** 在某些字体中使用时非常令人困惑，对于有视觉问题的人来说甚至更难区分。
- 删除标点字符可防止换行符混淆。
- 双击会将整数作为一个单词（如果全是字母数字）选择为一个单词。



```go
package shortener

import (
	"crypto/sha256"
	"fmt"
	"github.com/itchyny/base58-go"
	"math/big"
	"os"
)

// 使用sha256Of()函数来计算输入字符串的SHA256哈希值。
// sha256可以将任何字符串转换为一个256位的哈希值。为了防止哈希碰撞，我们使用SHA256算法。
func sha256Of(input string) []byte {
	algorithm := sha256.New()
	algorithm.Write([]byte(input))
	return algorithm.Sum(nil)
}

// 为了使哈希值更短，我们使用base58Encoded()函数将其编码为Base58。
// base58可以去除容易混淆的字符，例如0和O，1和l，以及+和/。
func base58Encoded(bytes []byte) string {
	encoding := base58.BitcoinEncoding
	encoded, err := encoding.Encode(bytes)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	return string(encoded)
}

// GenerateShortLink函数将原始网址和用户ID作为输入，并返回一个短网址。
// userId可以防止哈希碰撞。如果两个用户使用相同的原始网址，我们将为每个用户生成不同的短网址。
func GenerateShortLink(initialLink string, userId string) string {
	// 使用SHA256哈希算法来计算原始网址和用户ID的哈希值。
	urlHashBytes := sha256Of(initialLink + userId)
	// 将哈希值转换为一个大整数。
	generatedNumber := new(big.Int).SetBytes(urlHashBytes).Uint64()
	// 将大整数转换为Base58编码的字符串。
	finalString := base58Encoded([]byte(fmt.Sprintf("%d", generatedNumber)))
	// 返回前8个字符。
	return finalString[:8]
}

```



### 三. 3.缩短器单元测试

```go
package shortener

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const UserId = "e0dba740-fc4b-4977-872c-d360239e6b1a"

func TestShortLinkGenerator(t *testing.T) {
	initialLink_1 := "https://opensource.tencent.com/summer-of-code"
	shortLink_1 := GenerateShortLink(initialLink_1, UserId)

	initialLink_2 := "https://opensource.alibaba.com/"
	shortLink_2 := GenerateShortLink(initialLink_2, UserId)

	initialLink_3 := "https://opensource.google/"
	shortLink_3 := GenerateShortLink(initialLink_3, UserId)

	assert.Equal(t, shortLink_1, "fSjjvszt")
	assert.Equal(t, shortLink_2, "GYw5AcQz")
	assert.Equal(t, shortLink_3, "EPz1wNJG")
}

```



## 四、转发（重定向）

### 四. 1.处理程序和port

在不浪费更多时间的情况下，让我们继续创建**处理程序**包并在其中定义我们的处理器函数。
创建一个名为 **handler** 的文件夹，并放入一个名为 .
之后，我们的项目目录应该看起来像下面的树：`handlers.go`

```bash
├── go.mod
├── go.sum
├── handler
│   └── handlers.go
├── main.go
├── shortener
│   ├── shorturl_generator.go
│   └── shorturl_generator_test.go
└── store
    ├── store_service.go
    └── store_service_test.go
```



### 四. 1.2. 实现

**步驟**1：我们将从实现'CreateShortUrl（）'处理程序函数开始，这应该非常简单：

- 我们将获取创建请求正文，解析它并提取初始长 url 和 userId。
- 调用我们在[三. 2.缩短器实现](#三. 2.缩短器实现)中实现的并生成我们的缩短哈希值()。`shortener.GenerateShortLink()`
- 最后将输出的映射与初始长 url 存储，在这里，我们将使用我们在[二. 2.Store API 设计和实施 ](#二. 2.Store API 设计和实施)实现的`hash/shortUrl``store.SaveUrlMapping()`

**步骤 2 ：**第二步也是最后一步是关于实现重定向处理程序，它将包括：`HandleShortUrlRedirect()`

- 从路径参数获取短网址`/:shortUrl`
- 调用存储以检索与路径中提供的短 URL 对应的初始 URL。
- 最后应用 http 重定向功能

```go
package handler

import (
	"github.com/Poldroc/go-url-shortener/shortener"
	"github.com/Poldroc/go-url-shortener/store"
	"github.com/gin-gonic/gin"
	"net/http"
)

type UrlCreationRequest struct {
	LongUrl string `json:"long_url" binding:"required"`
	UserId  string `json:"user_id" binding:"required"`
}

// CreateShortUrl 将长网址转换为短网址。
func CreateShortUrl(c *gin.Context) {
	var creationRequest UrlCreationRequest
	if err := c.ShouldBindJSON(&creationRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	shortUrl := shortener.GenerateShortLink(creationRequest.LongUrl, creationRequest.UserId)
	store.SaveUrlMapping(shortUrl, creationRequest.LongUrl, creationRequest.UserId)

	host := "http://localhost:8080/"
	c.JSON(200, gin.H{
		"message":   "short url created successfully",
		"short_url": host + shortUrl,
	})

}

// HandleShortUrlRedirect 将短网址重定向到原始网址。
func HandleShortUrlRedirect(c *gin.Context) {
	shortUrl := c.Param("shortUrl")
	initialUrl := store.RetrieveInitialUrl(shortUrl)
	c.Redirect(302, initialUrl)

}

```



在 `main.go` 中实现使用:

```go
package main

import (
	"fmt"
	"github.com/Poldroc/go-url-shortener/handler"
	"github.com/Poldroc/go-url-shortener/store"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to the URL Shortener API",
		})
	})

	r.POST("/create-short-url", func(c *gin.Context) {
		handler.CreateShortUrl(c)
	})

	r.GET("/:shortUrl", func(c *gin.Context) {
		handler.HandleShortUrlRedirect(c)
	})

	// 初始化Redis
	store.InitializeStore()

	err := r.Run(":8080")
	if err != nil {
		panic(fmt.Sprintf("Failed to start the web server - Error: %v", err))
	}
}

```





### 四. 2.测试

- **步骤1**：运行/启动项目（文件是入口点）
  服务器应该在[localhost：8808](http://localhost:8080)启动`main.go`

* **步骤 2**：请求 URL 缩短操作。我们可以将下面的请求正文post到指定的url。

  post到 http://localhost:8080/create-short-url

```json
{
    "long_url": "https://opensource.tencent.com/summer-of-code",
    "user_id" : "e0dba740-fc4b-4977-872c-d360239e6b10"
}
```

​      响应如下：

```json
{
    "message": "short url created successfully",
    "short_url": "http://localhost:8080/UNNKLoJm"
}
```

* **步骤 3**：访问short_url，即可重定向回原始长 URL

