package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/aws/aws-sdk-go-v2/config" // v2 SDK config
)

// DynamoDBのテーブル名は環境変数から取得
// グローバル変数としてAWS SDK v2クライアントとテーブル名を定義
var (
	dynamodbClient *dynamodb.Client
	s3Client       *s3.Client
	tableName      = os.Getenv("TABLE_NAME")
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
// "acc"や"pass@k"などキーが動的に変わるため、map[string]float64として定義します。
type TaskResult map[string]float64

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

// init関数はLambdaの起動時に一度だけ実行され、AWS SDKクライアントを初期化します。
func init() {
	// 環境変数 TABLE_NAME が設定されているか確認
	if tableName == "" {
		log.Fatalf("環境変数 TABLE_NAME が設定されていません。")
	}

	// AWS SDK設定のロード (リージョンは環境変数またはデフォルト設定から取得)
	// init関数ではリクエストコンテキストがないため context.TODO() を使用
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("ap-northeast-1"))
	if err != nil {
		log.Fatalf("AWS SDK設定の読み込みに失敗しました: %v", err)
	}

	// S3クライアントとDynamoDBクライアントの作成
	s3Client = s3.NewFromConfig(cfg)
	dynamodbClient = dynamodb.NewFromConfig(cfg)
	log.Println("AWS SDK v2クライアントが正常に初期化されました。")
}

func main() {
	lambda.Start(handler)
}

// AWS Lambdaのメイン関数
// S3 Putイベントを処理し、S3からJSONファイルを読み込み、パースしてベンチマーク結果をDynamoDBに保存
func handler(ctx context.Context, s3Event events.S3Event) error {
	if len(s3Event.Records) == 0 {
		log.Println("S3イベントレコードがありません。")
		return nil // 処理するレコードがない場合は正常終了
	}

	for _, record := range s3Event.Records {
		bucketName := record.S3.Bucket.Name
		objectKey := record.S3.Object.Key

		log.Printf("S3イベントを受信しました: バケット '%s', キー '%s'\n", bucketName, objectKey)

		// S3からJSONファイルを読み込む
		getObjectOutput, err := s3Client.GetObject(ctx, &s3.GetObjectInput{ // グローバルなs3Clientを使用
			Bucket: aws.String(bucketName),
			Key:    aws.String(objectKey),
		})
		if err != nil {
			return fmt.Errorf("S3オブジェクトの取得に失敗しました (バケット: %s, キー: %s): %w", bucketName, objectKey, err)
		}
		defer getObjectOutput.Body.Close() // 必ずBodyを閉じる

		jsonData, err := io.ReadAll(getObjectOutput.Body) // io.ReadAllを使用 (ioutil.ReadAllは非推奨)
		if err != nil {
			return fmt.Errorf("S3オブジェクトの読み込みに失敗しました: %w", err)
		}

		var resultFile ResultFile
		if err := json.Unmarshal(jsonData, &resultFile); err != nil {
			return fmt.Errorf("JSONのアンマーシャリングに失敗しました: %w", err)
		}

		log.Println("model_configのパースに成功しました。")
		log.Println("===================================================")

		modelConfig := resultFile.ConfigGeneral.ModelConfig
		log.Printf("モデル名: %s\n", modelConfig.ModelName)

		log.Println("\n--- 生成パラメータ ---")
		genParams := modelConfig.GenerationParameters
		log.Printf("  最大新規トークン数: %d\n", genParams.MaxNewTokens)
		log.Printf("  Temperature: %.1f\n", genParams.Temperature)
		log.Printf("  Top K: %d\n", genParams.TopK)
		log.Printf("  Top P: %.1f\n", genParams.TopP)

		log.Println("===================================================")

		log.Println("\n--- ベンチマーク結果 ---")
		for taskName, taskResult := range resultFile.Results {
			log.Printf("  タスク: %s\n", taskName)
			for metricName, value := range taskResult {
				log.Printf("    %s: %v\n", metricName, value)
			}
		}
		log.Println("===================================================")

		// DynamoDBにアイテムを保存
		log.Println("\nDynamoDBにベンチマーク結果を登録します...")
		item, err := saveBenchmarkToDB(ctx, &resultFile, objectKey) // グローバルなdynamodbClientとtableNameを使用
		if err != nil {
			return fmt.Errorf("DynamoDBへのアイテム登録に失敗しました: %w", err)
		}

		log.Printf("モデル '%s' のベンチマーク結果をID '%s' でテーブル '%s' に正常に登録しました。\n", item.ModelName, item.WorkflowID, tableName)
		log.Println("===================================================")
		
	}

	return nil // すべてのレコードの処理が正常に完了
}

// saveBenchmarkToDB は、パースされたベンチマーク結果をDynamoDBに保存します。
func saveBenchmarkToDB(ctx context.Context, resultFile *ResultFile, objectKey string) (*BenchmarkItem, error) { // dynamodbClientとtableNameを引数から削除
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

	// DynamoDBに保存するアイテムを作成
	benchmarkItem := &BenchmarkItem{
		WorkflowID:           workflowID,
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
	_, err = dynamodbClient.PutItem(ctx, &dynamodb.PutItemInput{ // グローバルなdynamodbClientを使用
		TableName: aws.String(tableName),
		Item:      av,
	})
	if err != nil {
		return nil, fmt.Errorf("DynamoDBへのPutItem呼び出しに失敗しました: %w", err)
	}

	return benchmarkItem, nil
}
