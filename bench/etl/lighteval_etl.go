package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/google/uuid"
)

// JSONファイル全体の構造に合わせて、最上位のキーを定義します。
type ResultFile struct {
	ConfigGeneral ConfigGeneral `json:"config_general"`
	Results       Results       `json:"results"`
	// config_tasks はキーが動的なため、map[string]interface{} で柔軟に受け取ります
	ConfigTasks map[string]interface{} `json:"config_tasks"`
}

// ConfigGeneral は "config_general" オブジェクトに対応します。
type ConfigGeneral struct {
	ModelConfig ModelConfig `json:"model_config"`
}

// Results は "results" オブジェクトに対応します。
// キーがタスク名で動的に変わるため、map[string]TaskResultとして定義します。
type Results map[string]TaskResult

// TaskResult は各タスクの結果を格納します。
// JSONのキーに`@`が含まれるため、jsonタグでフィールド名を指定します。
type TaskResult struct {
	PassAtK        float64 `json:"pass@k_with_k&n" dynamodbav:"pass_at_k"`
	PassAtKStdErr  float64 `json:"pass@k_with_k&n_stderr" dynamodbav:"pass_at_k_stderr"`
}

// ModelConfig は "model_config" オブジェクトに対応します。
type ModelConfig struct {
	ModelName            string               `json:"model_name"`
	GenerationParameters GenerationParameters `json:"generation_parameters"`
}

// GenerationParameters は "generation_parameters" オブジェクトに対応します。
type GenerationParameters struct {
	RepetitionPenalty float64 `json:"repetition_penalty" dynamodbav:"repetition_penalty"`
	FrequencyPenalty  float64 `json:"frequency_penalty" dynamodbav:"frequency_penalty"`
	MaxNewTokens      int     `json:"max_new_tokens" dynamodbav:"max_new_tokens"`
	Temperature       float64 `json:"temperature" dynamodbav:"temperature"`
	TopK              int     `json:"top_k" dynamodbav:"top_k"`
	MinP              float64 `json:"min_p" dynamodbav:"min_p"`
	TopP              float64 `json:"top_p" dynamodbav:"top_p"`
}

// Node はノード情報を表します。
type Node struct {
	GPUName   string `dynamodbav:"gpu_name"`
	GPUCount  int    `dynamodbav:"gpu_count"`
	NodeCount int    `dynamodbav:"node_count"`
}

// BenchmarkItem はDynamoDBに格納するアイテムの構造体です。
type BenchmarkItem struct {
	WorkflowID           string               `dynamodbav:"workflow_id"` // Partition Key
	Timestamp            int64                `dynamodbav:"timestamp"`   // Sort Key
	DateTime             string               `dynamodbav:"datetime"`
	ModelName            string               `dynamodbav:"model_name"`
	DatasetName          string               `dynamodbav:"dataset_name"`
	BenchmarkType        string               `dynamodbav:"benchmark_type"`
	Framework            string               `dynamodbav:"framework"`
	NodeInfo             Node                 `dynamodbav:"node_info"`
	// genai-perfの出力を名前を合わせる(profile)
	Profile              Results              `dynamodbav:"profile"`
	GenerationParameters GenerationParameters `dynamodbav:"generation_parameters"`
}



func main() {
	if len(os.Args) < 2 {
		fmt.Println("使用法: go run lighteval_etl.go <JSONファイルパス>")
		os.Exit(1)
	}
	jsonFilePath := os.Args[1]

	jsonData, err := os.ReadFile(jsonFilePath)
	if err != nil {
		log.Fatalf("JSONファイルの読み込みに失敗しました: %v", err)
	}

	var resultFile ResultFile

	if err := json.Unmarshal(jsonData, &resultFile); err != nil {
		log.Fatalf("JSONのアンマーシャリングに失敗しました: %v", err)
	}

	fmt.Println("model_configのパースに成功しました。")
	fmt.Println("===================================================")

	modelConfig := resultFile.ConfigGeneral.ModelConfig
	fmt.Printf("モデル名: %s\n", modelConfig.ModelName)


	fmt.Println("\n--- 生成パラメータ ---")
	genParams := modelConfig.GenerationParameters
	fmt.Printf("  最大新規トークン数: %d\n", genParams.MaxNewTokens)
	fmt.Printf("  Temperature: %.1f\n", genParams.Temperature)
	fmt.Printf("  Top K: %d\n", genParams.TopK)
	fmt.Printf("  Top P: %.1f\n", genParams.TopP)

	fmt.Println("===================================================")

	fmt.Println("\n--- ベンチマーク結果 ---")
	for taskName, taskResult := range resultFile.Results {
		fmt.Printf("  タスク: %s\n", taskName)
		fmt.Printf("    pass@k_with_k&n: %.4f (stderr: %.4f)\n", taskResult.PassAtK, taskResult.PassAtKStdErr)
	}

	fmt.Println("===================================================")

	// --- DynamoDBへの登録処理 ---
	fmt.Println("\nDynamoDBにベンチマーク結果を登録します...")

	// 1. AWS設定のロード (リージョンは東京を想定)
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("ap-northeast-1"))
	if err != nil {
		log.Fatalf("AWS SDK設定の読み込みに失敗しました: %v", err)
	}

	// 2. DynamoDBクライアントの作成
	dynamodbClient := dynamodb.NewFromConfig(cfg)
	tableName := "benchmark-result"

	// 3. DynamoDBにアイテムを保存
	item, err := saveBenchmarkToDB(context.TODO(), dynamodbClient, tableName, &resultFile)
	if err != nil {
		log.Fatalf("DynamoDBへのアイテム登録に失敗しました: %v", err)
	}

	fmt.Printf("モデル '%s' のベンチマーク結果をID '%s' でテーブル '%s' に正常に登録しました。\n", item.ModelName, item.WorkflowID, tableName)
	fmt.Println("===================================================")
	
}

// saveBenchmarkToDB は、パースされたベンチマーク結果をDynamoDBに保存します。
func saveBenchmarkToDB(ctx context.Context, dynamodbClient *dynamodb.Client, tableName string, resultFile *ResultFile) (*BenchmarkItem, error) {
	// TODO: ノード情報を動的に取得する。現在は固定値。
	nodeItem := Node{
		GPUName:   "NVIDIA-A100-SXM4-80GB",
		GPUCount:  1,
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

	// DynamoDBに保存するアイテムを作成
	benchmarkItem := &BenchmarkItem{
		WorkflowID:           uuid.NewString(), // UUIDを生成してパーティションキーとして使用
		Timestamp:            now.Unix(),       // Unixタイムスタンプ（秒）をソートキーとして使用
		DateTime:             now.In(jst).Format(time.RFC3339),
		ModelName:            modelName,
		DatasetName:          datasetName,
		BenchmarkType:        "accuracy",
		Framework:            "Lighteval",
		NodeInfo:             nodeItem,
		Profile:              resultFile.Results,
		GenerationParameters: resultFile.ConfigGeneral.ModelConfig.GenerationParameters,
	}

	// Goの構造体をDynamoDBの属性値マップに変換
	av, err := attributevalue.MarshalMap(benchmarkItem)
	if err != nil {
		return nil, fmt.Errorf("DynamoDBアイテムのマーシャリングに失敗しました: %w", err)
	}

	// PutItem APIの呼び出し
	_, err = dynamodbClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      av,
	})
	if err != nil {
		return nil, fmt.Errorf("DynamoDBへのPutItem呼び出しに失敗しました: %w", err)
	}

	return benchmarkItem, nil
}
