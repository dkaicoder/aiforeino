package rag

import (
	"context"
	"fmt"
	"main/internal/database"
	"main/pkg/llm"

	redisInd "github.com/cloudwego/eino-ext/components/indexer/redis"
	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
)

const (
	prefix = "OuterCyrex:"
	index  = "OuterIndex"
)

type Input struct {
	FilePath string `json:"file_path" jsonschema:"description=Absolute path to the uploaded document file"`
}
type Output struct {
	Answer  string   `json:"answer"`
	Sources []string `json:"sources"`
}

type scoreTask struct {
	Doc []*schema.Document
}

// scoredChunk is the per-chunk result produced by the inner BatchNode workflow.
type scoredChunk struct {
	Doc []*schema.Document
}

func BuildWorkflow(ctx context.Context) *compose.Workflow[Input, Output] {
	wf := compose.NewWorkflow[Input, Output]()

	// load: 读取文件
	wf.AddLambdaNode("load", compose.InvokableLambda(
		func(ctx context.Context, in Input) ([]*schema.Document, error) {
			load, err := llm.NewLoader(ctx)
			if err != nil {
				return nil, err
			}
			doc, err := load.Load(ctx, document.Source{
				URI: in.FilePath,
			})
			if err != nil {
				panic(err)
			}
			return doc, nil
		},
	)).AddInput(compose.START)

	// chunk: 分块
	wf.AddLambdaNode("chunk", compose.InvokableLambda(
		func(ctx context.Context, docs []*schema.Document) (scoreTask, error) {
			splitter, err := llm.NewSplitter(ctx)
			if err != nil {
				return scoreTask{}, err
			}
			docss, err := splitter.Transform(ctx, docs)
			if err != nil {
				panic(err)
			}
			ss := scoreTask{
				Doc: docss,
			}
			return ss, nil
		},
	)).AddInput("load")
	// score:
	wf.AddLambdaNode("score", compose.InvokableLambda(
		func(ctx context.Context, in scoreTask) ([]scoredChunk, error) {
			for _, d := range in.Doc {
				uuids, _ := uuid.NewUUID()
				d.ID = uuids.String()
			}
			s := []scoredChunk{
				{
					Doc: in.Doc,
				},
			}
			err := InitVectorIndex(ctx)
			if err != nil {
				return s, err
			}

			newIndexer, err := newIndexer(ctx)
			if err != nil {
				return s, err
			}
			_, err = newIndexer.Store(ctx, in.Doc)
			if err != nil {
				return s, err
			}
			return s, nil
		},
	)).AddInput("chunk")

	wf.AddLambdaNode("answer", compose.InvokableLambda(
		func(ctx context.Context, in []scoredChunk) (Output, error) {
			fmt.Println(in)
			s := Output{}
			return s, nil
		},
	)).AddInput("score")
	wf.End().AddInput("answer")

	return wf
}

func newIndexer(ctx context.Context) (*redisInd.Indexer, error) {
	r := database.RedisDb
	em, err := llm.NewEmbeddingFactory(ctx, "doubao-embedding-text-240715")
	if err != nil {
		return nil, err
	}
	i, err := redisInd.NewIndexer(ctx, &redisInd.IndexerConfig{
		Client:           r,
		KeyPrefix:        prefix,
		DocumentToHashes: nil,
		BatchSize:        10,
		Embedding:        em,
	})
	if err != nil {
		return nil, err
	}
	return i, nil
}

func InitVectorIndex(ctx context.Context) error {

	r := database.RedisDb

	_, _ = r.Do(ctx, "FT.DROPINDEX", index).Result()

	createIndexArgs := []interface{}{
		"FT.CREATE", index,
		"ON", "HASH",
		"PREFIX", "1", prefix,
		"SCHEMA",
		"content", "TEXT",
		"vector_content", "VECTOR", "FLAT",
		"6",
		"TYPE", "FLOAT32",
		"DIM", 2560,
		"DISTANCE_METRIC", "COSINE",
	}

	if err := r.Do(ctx, createIndexArgs...).Err(); err != nil {
		return err
	}

	if _, err := r.Do(ctx, "FT.INFO", index).Result(); err != nil {
		return err
	}
	infoCmd := r.Do(ctx, "FT.INFO", index)
	info, err := infoCmd.Result()
	if err != nil {
		panic(fmt.Sprintf("获取索引信息失败：%v", err))
	}
	fmt.Printf("索引信息：%v\n", info)
	return nil
}

//func scoreOneChunk(ctx context.Context, t scoreTask) (scoredChunk, error) {
//	for _, d := range t.Doc {
//		uuids, _ := uuid.NewUUID()
//		d.ID = uuids.String()
//	}
//	s := scoredChunk{
//		Doc: t.Doc,
//	}
//	return s, nil
//}

//func newScoreWorkflow(cm context.Context) *compose.Workflow[scoreTask, scoredChunk] {
//	wf := compose.NewWorkflow[scoreTask, scoredChunk]()
//	wf.AddLambdaNode("score_chunk", compose.InvokableLambda(
//		func(ctx context.Context, t scoreTask) (scoredChunk, error) {
//			return scoreOneChunk(ctx, t)
//		},
//	)).AddInput(compose.START)
//	wf.End().AddInput("score_chunk")
//	return wf
//}
