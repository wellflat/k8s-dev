// main.go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// Movie はDynamoDBテーブルのアイテムを表す構造体です。
// `dynamodbav` タグを使って、GoのフィールドとDynamoDBの属性をマッピングします。
type Movie struct {
	Year  int       `dynamodbav:"year"`
	Title string    `dynamodbav:"title"`
	Info  MovieInfo `dynamodbav:"info"`
}

// MovieInfo はネストされた情報を持つ構造体です。
type MovieInfo struct {
	Plot   string   `dynamodbav:"plot"`
	Rating float64  `dynamodbav:"rating"`
	Actors []string `dynamodbav:"actors"`
}

func main() {
	// --- 1. AWS設定のロード ---
	// 環境変数や ~/.aws/credentials ファイルから設定を自動で読み込みます。
	// リージョンは "ap-northeast-1" (東京) を指定しています。
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("ap-northeast-1"))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	// --- 2. DynamoDBクライアントの作成 ---
	svc := dynamodb.NewFromConfig(cfg)

	// --- 3. 登録するデータの作成 ---
	movie := Movie{
		Year:  2015,
		Title: "The Big Short",
		Info: MovieInfo{
			Plot:   "A group of investors bet against the US mortgage market.",
			Rating: 8.4,
			Actors: []string{"Christian Bale", "Steve Carell", "Ryan Gosling"},
		},
	}

	// --- 4. Goの構造体をDynamoDBの属性値マップに変換 ---
	av, err := attributevalue.MarshalMap(movie)
	if err != nil {
		log.Fatalf("failed to marshal movie, %v", err)
	}

	// --- 5. PutItem APIの呼び出し ---
	tableName := "movie"

	input := &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      av,
	}

	_, err = svc.PutItem(context.TODO(), input)
	if err != nil {
		log.Fatalf("failed to put item, %v", err)
	}

	fmt.Printf("Successfully added '%s' (%d) to table %s\n", movie.Title, movie.Year, tableName)
}
