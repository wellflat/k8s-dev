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
	P1   float64 `json:"p1" dynamodbav:"p1"`
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

// TelemetryValue は、テレメトリメトリクスの個々の値を表します。
type TelemetryValue struct {
	Avg float64 `json:"avg" dynamodbav:"avg"`
	P1  float64 `json:"p1" dynamodbav:"p1"`
	P5  float64 `json:"p5" dynamodbav:"p5"`
	P10 float64 `json:"p10" dynamodbav:"p10"`
	P25 float64 `json:"p25" dynamodbav:"p25"`
	P50 float64 `json:"p50" dynamodbav:"p50"`
	P75 float64 `json:"p75" dynamodbav:"p75"`
	P90 float64 `json:"p90" dynamodbav:"p90"`
	P95 float64 `json:"p95" dynamodbav:"p95"`
	P99 float64 `json:"p99" dynamodbav:"p99"`
	Min float64 `json:"min" dynamodbav:"min"`
	Max float64 `json:"max" dynamodbav:"max"`
	Std float64 `json:"std" dynamodbav:"std"`
}

// TelemetryMetric は、単位とGPUごとのテレメトリデータを表します。
type TelemetryMetric struct {
	Unit string                    `json:"unit" dynamodbav:"unit"`
	Data map[string]TelemetryValue `json:"-" dynamodbav:"data"` // JSONのキーを動的に扱う
}

type TelemetryStats struct {
	GpuPowerUsage  TelemetryMetric `json:"gpu_power_usage" dynamodbav:"gpu_power_usage"`
	GpuUtilization TelemetryMetric `json:"gpu_utilization" dynamodbav:"gpu_utilization"`
}

// InputConfig は、ベンチマークの入力設定を表します。
type InputConfig struct {
	ModelNames []string `json:"model_names" dynamodbav:"model_names"`
}

// ProfileExport は、JSONファイル全体の構造を表します。
type ProfileExport struct {
	RequestThroughput               Metric `json:"request_throughput" dynamodbav:"request_throughput"`
	RequestLatency                  LatencyMetric `json:"request_latency" dynamodbav:"request_latency"`
	RequestCount                    Metric        `json:"request_count" dynamodbav:"request_count"`
	TimeToFirstToken                LatencyMetric `json:"time_to_first_token" dynamodbav:"time_to_first_token"`
	TimeToSecondToken               LatencyMetric `json:"time_to_second_token" dynamodbav:"time_to_second_token"`
	InterTokenLatency               LatencyMetric `json:"inter_token_latency" dynamodbav:"inter_token_latency"`
	OutputTokenThroughput           Metric        `json:"output_token_throughput" dynamodbav:"output_token_throughput"`
	OutputTokenThroughputPerUser    LatencyMetric `json:"output_token_throughput_per_user" dynamodbav:"output_token_throughput_per_user"`
	OutputSequenceLength            LatencyMetric `json:"output_sequence_length" dynamodbav:"output_sequence_length"`
	InputSequenceLength             LatencyMetric `json:"input_sequence_length" dynamodbav:"input_sequence_length"`
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
	Telemetry TelemetryStats `dynamodbav:"gpu_telemetry"`
}

// UnmarshalJSON は TelemetryMetric のためのカスタムアンマーシャラーです。
// これにより、"unit" 以外の動的なキー（GPU ID）を `Data` マップに格納できます。
func (t *TelemetryMetric) UnmarshalJSON(data []byte) error {
	// 一時的なマップにすべてのキーと値をデコード
	var rawData map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawData); err != nil {
		return err
	}

	// "unit" フィールドを抽出し、マップから削除
	if unitValue, ok := rawData["unit"]; ok {
		if err := json.Unmarshal(unitValue, &t.Unit); err != nil {
			return fmt.Errorf("failed to unmarshal unit: %w", err)
		}
		delete(rawData, "unit")
	}

	// 残りのデータ（GPUごとのデータ）を `Data` フィールドにデコード
	// まずマップを再エンコードし、それを `t.Data` にデコードする
	remainingData, err := json.Marshal(rawData)
	if err != nil {
		return fmt.Errorf("failed to marshal remaining data: %w", err)
	}
	return json.Unmarshal(remainingData, &t.Data)
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

	// パースしたデータを保持するための変数を定義します。
	// ProfileExportはパフォーマンスメトリクスのみを保持します。
	var profileData ProfileExport
	// InputConfigとTelemetryStatsは別の専用の構造体でパースします。
	var inputConfig struct {
		InputConfig InputConfig `json:"input_config"`
	}
	var telemetryStats struct {
		TelemetryStats TelemetryStats `json:"telemetry_stats"`
	}

	// JSONデータを構造体にアンマーシャリング（デコード）
	if err := json.Unmarshal(jsonData, &profileData); err != nil {
		log.Fatalf("json unmarshaling error: %v", err)
	}
	if err := json.Unmarshal(jsonData, &inputConfig); err != nil {
		log.Fatalf("json unmarshaling error for input_config: %v", err)
	}
	if err := json.Unmarshal(jsonData, &telemetryStats); err != nil {
		log.Fatalf("json unmarshaling error for telemetry_stats: %v", err)
	}

	// パースしたデータの一部をコンソールに出力して確認
	fmt.Println("genai-perf プロファイルのエクスポートファイルのパースに成功しました。")
	fmt.Println("===================================================")
	fmt.Printf("モデル: %s\n", inputConfig.InputConfig.ModelNames[0])
	fmt.Println("---------------------------------------------------")
	fmt.Printf("リクエストスループット: %.2f %s\n", profileData.RequestThroughput.Avg, profileData.RequestThroughput.Unit)
	fmt.Printf("リクエストレイテンシ (平均): %.2f %s\n", profileData.RequestLatency.Avg, profileData.RequestLatency.Unit)
	fmt.Printf("最初のトークンまでの時間 (平均): %.2f %s\n", profileData.TimeToFirstToken.Avg, profileData.TimeToFirstToken.Unit)
	fmt.Printf("トークン間レイテンシ (平均): %.2f %s\n", profileData.InterTokenLatency.Avg, profileData.InterTokenLatency.Unit)
	fmt.Printf("出力トークンスループット (平均): %.2f %s\n", profileData.OutputTokenThroughput.Avg, profileData.OutputTokenThroughput.Unit)
	fmt.Println("===================================================")
	fmt.Printf("テレメトリデータ (GPU Power Usage - %s):\n", telemetryStats.TelemetryStats.GpuPowerUsage.Unit)
	for gpuID, data := range telemetryStats.TelemetryStats.GpuPowerUsage.Data {
		fmt.Printf("  GPU %s (Avg): %.2f\n", gpuID, data.Avg)
	}
	fmt.Println("---------------------------------------------------")
	fmt.Printf("テレメトリデータ (GPU Utilization - %s):\n", telemetryStats.TelemetryStats.GpuUtilization.Unit)
	for gpuID, data := range telemetryStats.TelemetryStats.GpuUtilization.Data {
		fmt.Printf("  GPU %s (Avg): %.2f\n", gpuID, data.Avg)
	}

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

	item, err := saveBenchmarkToDB(context.TODO(), dynamodbClient, tableName, &profileData, &inputConfig.InputConfig, &telemetryStats.TelemetryStats)
	if err != nil {
		log.Fatalf("DynamoDBへのアイテム登録に失敗しました: %v", err)
	}

	fmt.Printf("モデル '%s' のベンチマーク結果をID '%s' でテーブル '%s' に正常に登録しました。\n", item.ModelName, item.WorkflowID, tableName)
	fmt.Println("===================================================")

}

// saveBenchmarkToDB は、パースされたベンチマーク結果をDynamoDBに保存します。
func saveBenchmarkToDB(ctx context.Context, dynamodbClient *dynamodb.Client, tableName string, profileData *ProfileExport, inputConfig *InputConfig, telemetryStats *TelemetryStats) (*BenchmarkItem, error) {
	// TODO: ノード情報を取得, 現在は固定値を入れている
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
		ModelName:     inputConfig.ModelNames[0],
		DatasetName:   "Open-Orca/OpenOrca",
		NodeInfo:      *nodeItem,
		DateTime:      now.In(jst).Format(time.RFC3339), // ISO 8601形式のJSTタイムスタンプ
		Timestamp:     now.Unix(),                       // Unixタイムスタンプ（秒）
		BenchmarkType: "inference",
		Profile:      *profileData,
		Telemetry:     *telemetryStats,
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
