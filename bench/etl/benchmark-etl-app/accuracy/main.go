package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/config" // v2 SDK config
)

// DynamoDBのテーブル名は環境変数から取得
// グローバル変数としてAWS SDK v2クライアントとテーブル名を定義
var (
	dynamodbClient *dynamodb.Client
	s3Client       *s3.Client
	tableName      = os.Getenv("TABLE_NAME")
)

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
		log.Printf("  総評価時間(秒): %s\n", resultFile.ConfigGeneral.TotalEvaluationTimeSecondes)

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
