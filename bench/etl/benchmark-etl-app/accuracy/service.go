package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// BenchmarkService はベンチマーク関連のビジネスロジックを扱うサービスです。
type BenchmarkService struct {
	repo BenchmarkRepository
}

// NewBenchmarkService は新しいBenchmarkServiceインスタンスを作成します。
func NewBenchmarkService(repo BenchmarkRepository) *BenchmarkService {
	return &BenchmarkService{repo: repo}
}

// CreateBenchmarkFromS3Result はS3の結果ファイルからベンチマークデータを作成し、永続化します。
func (s *BenchmarkService) CreateBenchmarkFromS3Result(ctx context.Context, resultFile *ResultFile, objectKey string) (*BenchmarkItem, error) {
	// TODO: ノード情報を動的に取得する。現在は固定値。
	nodeItem := Node{
		GPUName:   "NVIDIA-A100-SXM4-80GB",
		GPUCount:  4,
		NodeCount: 1,
	}

	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return nil, fmt.Errorf("JSTタイムゾーンの読み込みに失敗しました: %w", err)
	}
	now := time.Now()

	var datasetName string
	if len(resultFile.ConfigTasks) > 0 {
		for _, taskConfigRaw := range resultFile.ConfigTasks {
			if taskConfig, ok := taskConfigRaw.(map[string]interface{}); ok {
				if hfRepo, ok := taskConfig["hf_repo"].(string); ok {
					datasetName = hfRepo
					break
				}
			}
		}
	}

	modelName := strings.Replace(resultFile.ConfigGeneral.ModelConfig.ModelName, "openai/openai/", "openai/", 1)

	workflowID := uuid.NewString()
	parts := strings.SplitN(objectKey, "/", 2)
	if len(parts) > 0 && parts[0] != "" {
		workflowID = parts[0]
	}

	totalEvaluationTime, err := strconv.ParseFloat(resultFile.ConfigGeneral.TotalEvaluationTimeSecondes, 64)
	if err != nil {
		return nil, fmt.Errorf("total_evaluation_time_secondes のパースに失敗しました: %w", err)
	}

	benchmarkItem := &BenchmarkItem{
		WorkflowID:           workflowID,
		Timestamp:            now.Unix(),
		DateTime:             now.In(jst).Format(time.RFC3339),
		ModelName:            modelName,
		DatasetName:          datasetName,
		BenchmarkType:        "accuracy",
		Framework:            "Lighteval",
		NodeInfo:             nodeItem,
		Profile:              resultFile.Results,
		GenerationParameters: resultFile.ConfigGeneral.ModelConfig.GenerationParameters,
		TotalEvaluationTime:  totalEvaluationTime,
	}

	if err := s.repo.SaveBenchmark(ctx, benchmarkItem); err != nil {
		return nil, err
	}

	return benchmarkItem, nil
}