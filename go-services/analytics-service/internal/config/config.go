package config

import "github.com/zeromicro/go-zero/rest"

// Config 配置
type Config struct {
	rest.RestConf

	// MySQL配置
	Mysql MysqlConfig

	// Kafka配置
	Kafka KafkaConfig
}

// MysqlConfig MySQL配置
type MysqlConfig struct {
	DataSource string
}

// KafkaConfig Kafka配置
type KafkaConfig struct {
	Brokers []string // Kafka brokers列表
	Topic   string   // 消费的Topic
	GroupID string   // 消费者组ID
}
