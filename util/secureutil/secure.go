package secureutil

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"fmt"
)

func HmacSHA256(plaintext string, key string) string {
	hash := hmac.New(sha256.New, []byte(key)) // 创建哈希算法
	hash.Write([]byte(plaintext))             // 写入数据
	return fmt.Sprintf("%X", hash.Sum(nil))
}

func HmacMD5(plaintext string, key string) string {
	hash := hmac.New(md5.New, []byte(key)) // 创建哈希算法
	hash.Write([]byte(plaintext))          // 写入数据
	return fmt.Sprintf("%X", hash.Sum(nil))
}
