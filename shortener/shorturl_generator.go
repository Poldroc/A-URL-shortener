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
