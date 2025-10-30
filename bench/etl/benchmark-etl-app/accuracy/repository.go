package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// BenchmarkRepository はベンチマークデータを永続化するためのインターフェースです。
type BenchmarkRepository interface {
	SaveBenchmark(ctx context.Context, item *BenchmarkItem) error
}

// DynamoDBRepository はDynamoDBをデータストアとして使用するBenchmarkRepositoryの実装です。
type DynamoDBRepository struct {
	client    *dynamodb.Client
	tableName string
}

// NewDynamoDBRepository は新しいDynamoDBRepositoryインスタンスを作成します。
func NewDynamoDBRepository(client *dynamodb.Client, tableName string) *DynamoDBRepository {
	return &DynamoDBRepository{client: client, tableName: tableName}
}

// SaveBenchmark はBenchmarkItemをDynamoDBに保存します。
func (r *DynamoDBRepository) SaveBenchmark(ctx context.Context, item *BenchmarkItem) error {
	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("DynamoDBアイテムのマーシャリングに失敗しました: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{TableName: aws.String(r.tableName), Item: av})
	if err != nil {
		return fmt.Errorf("DynamoDBへのPutItem呼び出しに失敗しました: %w", err)
	}
	return nil
}