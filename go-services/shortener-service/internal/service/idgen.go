package service

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/bwmarrin/snowflake"
)

// IDGenerator ID生成器接口
type IDGenerator interface {
	GenerateID() (int64, error)
	GenerateShortCode() (string, error)
}

// SnowflakeIDGen 雪花算法ID生成器
type SnowflakeIDGen struct {
	node *snowflake.Node
}

// NewSnowflakeIDGen 创建雪花算法ID生成器
func NewSnowflakeIDGen(machineID int64) (*SnowflakeIDGen, error) {
	node, err := snowflake.NewNode(machineID)
	if err != nil {
		return nil, fmt.Errorf("failed to create snowflake node: %w", err)
	}
	return &SnowflakeIDGen{node: node}, nil
}

// GenerateID 生成唯一ID
func (g *SnowflakeIDGen) GenerateID() (int64, error) {
	return g.node.Generate().Int64(), nil
}

// GenerateShortCode 生成短链码
func (g *SnowflakeIDGen) GenerateShortCode() (string, error) {
	id := g.node.Generate().Int64()
	return encodeBase62(id), nil
}

// encodeBase62 将int64编码为Base62字符串
func encodeBase62(num int64) string {
	if num == 0 {
		return "0"
	}

	const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	result := ""
	n := num

	for n > 0 {
		remainder := n % 62
		result = string(base62Chars[remainder]) + result
		n = n / 62
	}

	return result
}

// GenerateRandomCode 生成指定长度的随机短链码
func GenerateRandomCode(length int) (string, error) {
	const charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	result := make([]byte, length)

	for i := range result {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		result[i] = charset[num.Int64()]
	}

	return string(result), nil
}
