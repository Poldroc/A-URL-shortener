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
