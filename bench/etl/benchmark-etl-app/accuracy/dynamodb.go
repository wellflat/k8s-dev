package main

import (
	"context"
	"strconv"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/google/uuid"
)

// saveBenchmarkToDB は、パースされたベンチマーク結果をDynamoDBに保存します。
func saveBenchmarkToDB(ctx context.Context, resultFile *ResultFile, objectKey string) (*BenchmarkItem, error) {
	// TODO: ノード情報を動的に取得する。現在は固定値。
	nodeItem := Node{
		GPUName:   "NVIDIA-A100-SXM4-80GB",
		GPUCount:  4,
		NodeCount: 1,
	}

	// JSTタイムゾーンを取得
	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return nil, fmt.Errorf("JSTタイムゾーンの読み込みに失敗しました: %w", err)
	}
	now := time.Now()

	// resultFile.ConfigTasksからデータセット名を取得
	var datasetName string
	if len(resultFile.ConfigTasks) > 0 {
		// 最初のタスク設定を取得
		for _, taskConfigRaw := range resultFile.ConfigTasks {
			if taskConfig, ok := taskConfigRaw.(map[string]interface{}); ok {
				if hfRepo, ok := taskConfig["hf_repo"].(string); ok {
					datasetName = hfRepo
				}
				if datasetName != "" {
					break // データセット名が見つかったらループを抜ける
				}
			}
		}
	}

	// ModelNameを整形する
	modelName := resultFile.ConfigGeneral.ModelConfig.ModelName
	modelName = strings.Replace(modelName, "openai/openai/", "openai/", 1)

	// S3のオブジェクトキーからWorkflowIDを抽出
	workflowID := uuid.NewString() // デフォルトはUUID
	parts := strings.SplitN(objectKey, "/", 2)
	if len(parts) > 0 && parts[0] != "" {
		workflowID = parts[0]
	}

	// total_evaluation_time_secondes (string) を float64 に変換
	totalEvaluationTime, err := strconv.ParseFloat(resultFile.ConfigGeneral.TotalEvaluationTimeSecondes, 64)
	if err != nil {
		return nil, fmt.Errorf("total_evaluation_time_secondes のパースに失敗しました: %w", err)
	}

	// DynamoDBに保存するアイテムを作成
	benchmarkItem := &BenchmarkItem{
		WorkflowID:           workflowID,
		Timestamp:            now.Unix(), // Unixタイムスタンプ（秒）をソートキーとして使用
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

	// Goの構造体をDynamoDBの属性値マップに変換
	av, err := attributevalue.MarshalMap(benchmarkItem)
	if err != nil {
		return nil, fmt.Errorf("DynamoDBアイテムのマーシャリングに失敗しました: %w", err)
	}

	// PutItem APIの呼び出し
	_, err = dynamodbClient.PutItem(ctx, &dynamodb.PutItemInput{ // グローバルなdynamodbClientを使用
		TableName: aws.String(tableName),
		Item:      av,
	})
	if err != nil {
		return nil, fmt.Errorf("DynamoDBへのPutItem呼び出しに失敗しました: %w", err)
	}

	return benchmarkItem, nil
}
