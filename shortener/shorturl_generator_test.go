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
