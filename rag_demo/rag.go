package rag_demo

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino-ext/components/document/loader/file"
	embedding "github.com/cloudwego/eino-ext/components/embedding/ark"
	redisInd "github.com/cloudwego/eino-ext/components/indexer/redis"
	"github.com/cloudwego/eino-ext/components/model/ark"
	redisRet "github.com/cloudwego/eino-ext/components/retriever/redis"
	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
	"github.com/redis/go-redis/v9"
)

type RAGEngine struct {
	indexName string
	prefix    string
	dimension int

	redis    *redis.Client
	embedder *embedding.Embedder

	Err error

	Loader    *file.FileLoader
	Splitter  document.Transformer
	Retriever *redisRet.Retriever
	Indexer   *redisInd.Indexer
	ChatModel *ark.ChatModel
}

func InitRAGEngine(ctx context.Context, index string, prefix string) (*RAGEngine, error) {
	r, err := initRAGEngine(ctx, index, prefix)
	if err != nil {
		return nil, err
	}

	r.newLoader(ctx)
	r.newSplitter(ctx)
	r.newIndexer(ctx)
	r.newRetriever(ctx)
	//r.newChatModel(ctx)

	return r, nil
}

func initRAGEngine(ctx context.Context, index string, prefix string) (*RAGEngine, error) {

	embedder, err := embedding.NewEmbedder(ctx, &embedding.EmbeddingConfig{
		APIKey: "358ad2c2-9d3b-4990-92c7-117cf25fdae3",
		Model:  "doubao-embedding-text-240715",
	})
	if err != nil {
		return nil, err
	}

	return &RAGEngine{
		indexName: index,
		prefix:    prefix,
		dimension: 2560,

		redis: redis.NewClient(&redis.Options{
			Addr:          fmt.Sprintf("%s:%d", "127.0.0.1", 6379),
			Protocol:      2,
			UnstableResp3: true,
		}),
		embedder: embedder,

		Loader:    nil,
		Splitter:  nil,
		Retriever: nil,
		Indexer:   nil,
		ChatModel: nil,
	}, nil
}

var systemPrompt = `
# Role: 你是数据导出助手
# Language: 中文
# Core Rules:
- 第一步：先解析用户需求中是否包含「明确的关联补充要求」：
  🔴 明确关联需求（直接生成连表SQL，不反问）：
  1. 用户提到“含/包含/附带/关联 + 关联字段对应信息”（如“含对应用户的用户名”、“附带gift_id对应的礼物名称”）；
  2. 用户直接要求“关联user表”“JOIN user表”等；
  🟢 模糊需求（先生成基础SQL，再反问）：
  1. 用户仅说“导出[表名]”（如“导出用户礼物记录表”）；
  2. 用户指定字段但未提关联（如“导出uid和礼物名称”）；
  🟡 无关联需求（直接生成基础SQL，不反问）：
  - 用户明确说“只导出[非关联字段]”（如“只导出礼物名称”）；
- 第二步：SQL生成规则：
  1. 明确关联需求：直接生成JOIN关联表的SQL，只SELECT用户指定/隐含的字段（禁止SELECT *）；
  2. 模糊需求：先生成基础SQL，再反问（列出需关联字段）；
  3. 无关联需求：只生成原表字段的SQL；
- 关联字段定义：
  1. 特殊关联字段：uid → 关联user表的id字段，可补充nickname（用户名）；
  2. 通用关联字段：所有以_id为后缀的字段（如gift_id、player_id）；
- 反问规则（仅模糊需求触发）：
  语言简洁，只列当前表的关联字段，如：“你提到的user_player表包含需关联的字段：uid（关联user表的用户ID），是否需要补充对应的用户名信息？”
以下是为你检索到的相关文档内容：
==== doc start ====
{documents}
==== doc end ====
`

func (r *RAGEngine) Generate(ctx context.Context, query string) (*schema.StreamReader[*schema.Message], error) {
	re, _ := redisRet.NewRetriever(ctx, &redisRet.RetrieverConfig{
		Client:            r.redis,
		Index:             r.indexName,
		VectorField:       "vector_content",
		DistanceThreshold: nil,
		Dialect:           2,
		ReturnFields:      []string{"content"},
		DocumentConverter: nil,
		TopK:              5,
		Embedding:         r.embedder,
	})

	query = "trajectories"
	docs, err := re.Retrieve(ctx, query)
	if err != nil {
		return nil, err
	}

	var docsContent string
	for _, doc := range docs {
		docsContent += doc.Content + "\n"
	}

	tpl := prompt.FromMessages(schema.FString,
		schema.SystemMessage(systemPrompt),
		schema.UserMessage(`问题: {content}`))

	messages, err := tpl.Format(ctx, map[string]any{
		"documents": docsContent,
		"content":   query,
	})
	if err != nil {
		return nil, err
	}
	return r.ChatModel.Stream(ctx, messages)
}
