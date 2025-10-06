package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

// JSONファイル全体の構造に合わせて、最上位のキーを定義します。
type ResultFile struct {
	ConfigGeneral ConfigGeneral `json:"config_general"`
	Results       Results       `json:"results"`
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
	PassAtK float64 `json:"pass@k_with_k&n"`
}

// ModelConfig は "model_config" オブジェクトに対応します。
type ModelConfig struct {
	ModelName            string               `json:"model_name"`
	GenerationParameters GenerationParameters `json:"generation_parameters"`
}

// GenerationParameters は "generation_parameters" オブジェクトに対応します。
type GenerationParameters struct {
	RepetitionPenalty float64 `json:"repetition_penalty"`
	FrequencyPenalty  float64 `json:"frequency_penalty"`
	MaxNewTokens      int     `json:"max_new_tokens"`
	Temperature       float64 `json:"temperature"`
	TopK              int     `json:"top_k"`
	MinP              float64 `json:"min_p"`
	TopP              float64 `json:"top_p"`
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
		fmt.Printf("    pass@k_with_k&n: %.4f\n", taskResult.PassAtK)
	}

	fmt.Println("===================================================")
}
