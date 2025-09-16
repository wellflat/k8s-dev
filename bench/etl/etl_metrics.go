package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/google/uuid"
)

// Metric は、単位と平均値を持つ単純なメトリクスを表します。
type Metric struct {
	Unit string  `json:"unit" dynamodbav:"unit"`
	Avg  float64 `json:"avg" dynamodbav:"avg"`
}

// LatencyMetric は、パーセンタイルを含む詳細なレイテンシやスループットのメトリクスを表します。
// この構造体は、同様の構造を持つ他の多くのメトリクスにも再利用できます。
type LatencyMetric struct {
	Unit string  `json:"unit" dynamodbav:"unit"`
	Avg  float64 `json:"avg" dynamodbav:"avg"`
	P25  float64 `json:"p25" dynamodbav:"p25"`
	P50  float64 `json:"p50" dynamodbav:"p50"`
	P75  float64 `json:"p75" dynamodbav:"p75"`
	P90  float64 `json:"p90" dynamodbav:"p90"`
	P95  float64 `json:"p95" dynamodbav:"p95"`
	P99  float64 `json:"p99" dynamodbav:"p99"`
	Min  float64 `json:"min" dynamodbav:"min"`
	Max  float64 `json:"max" dynamodbav:"max"`
	Std  float64 `json:"std" dynamodbav:"std"`
}

// InputConfig は、ベンチマークの入力設定を表します。
type InputConfig struct {
	Subcommand               string   `json:"subcommand" dynamodbav:"subcommand"`
	Model                    []string `json:"model" dynamodbav:"model"`
	Backend                  string   `json:"backend" dynamodbav:"backend"`
	Endpoint                 string   `json:"endpoint" dynamodbav:"endpoint"`
	Streaming                bool     `json:"streaming" dynamodbav:"streaming"`
	URL                      string   `json:"u" dynamodbav:"u"`
	OutputTokensMean         int      `json:"output_tokens_mean" dynamodbav:"output_tokens_mean"`
	RequestCount             int      `json:"request_count" dynamodbav:"request_count"`
	SyntheticInputTokensMean int      `json:"synthetic_input_tokens_mean" dynamodbav:"synthetic_input_tokens_mean"`
	Concurrency              int      `json:"concurrency" dynamodbav:"concurrency"`
	FormattedModelName       string   `json:"formatted_model_name" dynamodbav:"formatted_model_name"`
}


// ProfileExport は、JSONファイル全体の構造を表します。
type ProfileExport struct {
	RequestThroughput               Metric        `json:"request_throughput" dynamodbav:"request_throughput"`
	RequestLatency                  LatencyMetric `json:"request_latency" dynamodbav:"request_latency"`
	RequestCount                    Metric        `json:"request_count" dynamodbav:"request_count"`
	TimeToFirstToken                LatencyMetric `json:"time_to_first_token" dynamodbav:"time_to_first_token"`
	TimeToSecondToken               LatencyMetric `json:"time_to_second_token" dynamodbav:"time_to_second_token"`
	InterTokenLatency               LatencyMetric `json:"inter_token_latency" dynamodbav:"inter_token_latency"`
	OutputTokenThroughput           Metric        `json:"output_token_throughput" dynamodbav:"output_token_throughput"`
	OutputTokenThroughputPerRequest LatencyMetric `json:"output_token_throughput_per_request" dynamodbav:"output_token_throughput_per_request"`
	OutputSequenceLength            LatencyMetric `json:"output_sequence_length" dynamodbav:"output_sequence_length"`
	InputSequenceLength             LatencyMetric `json:"input_sequence_length" dynamodbav:"input_sequence_length"`
	InputConfig                     InputConfig   `json:"input_config" dynamodbav:"input_config"`
}

type Node struct {
	GPUName string `json:"gpu_name" dynamodbav:"gpu_name"`
	GPUCount int `json:"gpu_count" dynamodbav:"gpu_count"`
	NodeCount int `json:"node_count" dynamodbav:"node_count"`
}

// BenchmarkItem は、DynamoDBの'benchmark'テーブルに格納するアイテムの構造体です。
type BenchmarkItem struct {
	WorkflowID string `dynamodbav:"workflow_id"` // Partition Key
	ModelName string `dynamodbav:"model_name"`
	DatasetName string `dynamodbav:"dataset_name"`
	NodeInfo Node `dynamodbav:"node_info"`
	DateTime string `dynamodbav:"datetime"`
	Timestamp int64 `dynamodbav:"timestamp"` // Sort Key
	BenchmarkType string `dynamodbav:"benchmark_type"`
	// ProfileExport の全データをネストされたマップとして格納します
	Profile ProfileExport `dynamodbav:"profile"`
}

func main() {
	// コマンドライン引数としてファイルパスが渡されているかチェック
	if len(os.Args) < 2 {
		fmt.Println("usage: go run etl_metrics.go <json file path>")
		os.Exit(1)
	}
	jsonFilePath := os.Args[1]

	// JSONファイルを読み込む
	jsonData, err := os.ReadFile(jsonFilePath)
	if err != nil {
		log.Fatalf("json load error: %v", err)
	}

	// パースしたデータを保持するための変数を定義
	var profileData ProfileExport

	// JSONデータを構造体にアンマーシャリング（デコード）
	if err := json.Unmarshal(jsonData, &profileData); err != nil {
		log.Fatalf("json unmarshaling error: %v", err)
	}

	// パースしたデータの一部をコンソールに出力して確認
	fmt.Println("genai-perf プロファイルのエクスポートファイルのパースに成功しました。")
	fmt.Println("===================================================")
	fmt.Printf("モデル: %s\n", profileData.InputConfig.FormattedModelName)
	fmt.Printf("バックエンド: %s\n", profileData.InputConfig.Backend)
	fmt.Println("---------------------------------------------------")
	fmt.Printf("リクエストスループット: %.2f %s\n", profileData.RequestThroughput.Avg, profileData.RequestThroughput.Unit)
	fmt.Printf("リクエストレイテンシ (平均): %.2f %s\n", profileData.RequestLatency.Avg, profileData.RequestLatency.Unit)
	fmt.Printf("最初のトークンまでの時間 (平均): %.2f %s\n", profileData.TimeToFirstToken.Avg, profileData.TimeToFirstToken.Unit)
	fmt.Printf("トークン間レイテンシ (平均): %.2f %s\n", profileData.InterTokenLatency.Avg, profileData.InterTokenLatency.Unit)
	fmt.Printf("出力トークンスループット (平均): %.2f %s\n", profileData.OutputTokenThroughput.Avg, profileData.OutputTokenThroughput.Unit)
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

	item, err := saveBenchmarkToDB(context.TODO(), dynamodbClient, tableName, &profileData)
	if err != nil {
		log.Fatalf("DynamoDBへのアイテム登録に失敗しました: %v", err)
	}

	fmt.Printf("モデル '%s' のベンチマーク結果をID '%s' でテーブル '%s' に正常に登録しました。\n", item.ModelName, item.WorkflowID, tableName)
	fmt.Println("===================================================")
	
}

// saveBenchmarkToDB は、パースされたベンチマーク結果をDynamoDBに保存します。
func saveBenchmarkToDB(ctx context.Context, dynamodbClient *dynamodb.Client, tableName string, profileData *ProfileExport) (*BenchmarkItem, error) {
	nodeItem := &Node{
		GPUName: "NVIDIA-A100-SXM4-80GB",
		GPUCount: 1,
		NodeCount: 1,
	}
	// JSTタイムゾーンを取得
	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return nil, fmt.Errorf("JSTタイムゾーンの読み込みに失敗しました: %w", err)
	}
	now := time.Now()
	benchmarkItem := &BenchmarkItem{
		WorkflowID:    uuid.NewString(), // UUIDを生成してパーティションキーとして使用
		ModelName:     profileData.InputConfig.FormattedModelName,
		DatasetName:   "Open-Orca/OpenOrca",
		NodeInfo:      *nodeItem,
		DateTime:      now.In(jst).Format(time.RFC3339), // ISO 8601形式のJSTタイムスタンプ
		Timestamp:     now.Unix(),                       // Unixタイムスタンプ（秒）
		BenchmarkType: "inference",
		Profile:      *profileData,
	}

	// Goの構造体をDynamoDBの属性値マップに変換
	av, err := attributevalue.MarshalMap(benchmarkItem)
	if err != nil {
		return nil, fmt.Errorf("DynamoDBアイテムのマーシャリングに失敗しました: %w", err)
	}

	// PutItem APIの呼び出し
	input := &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      av,
	}

	if _, err = dynamodbClient.PutItem(ctx, input); err != nil {
		return nil, fmt.Errorf("DynamoDBへのPutItem呼び出しに失敗しました: %w", err)
	}

	return benchmarkItem, nil
}
