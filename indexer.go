package main

import (
	"context"
	"fmt"

	redisInd "github.com/cloudwego/eino-ext/components/indexer/redis"
)

func (r *RAGEngine) newIndexer(ctx context.Context) {
	i, err := redisInd.NewIndexer(ctx, &redisInd.IndexerConfig{
		Client:           r.redis,
		KeyPrefix:        r.prefix,
		DocumentToHashes: nil,
		BatchSize:        10,
		Embedding:        r.embedder,
	})
	if err != nil {
		r.Err = err
	}
	r.Indexer = i
}

func (r *RAGEngine) InitVectorIndex(ctx context.Context) error {
	_, _ = r.redis.Do(ctx, "FT.DROPINDEX", r.indexName).Result()

	createIndexArgs := []interface{}{
		"FT.CREATE", r.indexName,
		"ON", "HASH",
		"PREFIX", "1", r.prefix,
		"SCHEMA",
		"content", "TEXT",
		"vector_content", "VECTOR", "FLAT",
		"6",
		"TYPE", "FLOAT32",
		"DIM", r.dimension,
		"DISTANCE_METRIC", "COSINE",
	}

	if err := r.redis.Do(ctx, createIndexArgs...).Err(); err != nil {
		return err
	}

	if _, err := r.redis.Do(ctx, "FT.INFO", r.indexName).Result(); err != nil {
		return err
	}

	// 写入文档后，立即检查索引
	infoCmd := r.redis.Do(ctx, "FT.INFO", r.indexName)
	info, err := infoCmd.Result()
	if err != nil {
		panic(fmt.Sprintf("获取索引信息失败：%v", err))
	}
	// 打印索引信息，重点看“Number of docs”
	fmt.Printf("索引信息：%v\n", info)

	return nil
}
