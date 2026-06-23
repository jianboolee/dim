package main

import (
	"d-im/internal/bootstrap"
	"d-im/internal/config"
	"log"
)

// 数据库迁移
func main() {

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// 初始化索引
	if err := bootstrap.InitIndexes(cfg); err != nil {
		log.Fatal("Failed to init indexes:", err)
	}

	// 创建默认用户
	if err := bootstrap.InitSeed(cfg); err != nil {
		log.Fatal("Failed to init seed:", err)
	}
}
