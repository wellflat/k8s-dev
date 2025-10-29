package main

// JSONファイル全体の構造に合わせて、最上位のキーを定義します。
type ResultFile struct {
	ConfigGeneral ConfigGeneral `json:"config_general"`
	Results       Results       `json:"results"`
	// config_tasks はキーが動的なため、map[string]interface{} で柔軟に受け取ります
	ConfigTasks map[string]interface{} `json:"config_tasks"`
}

// ConfigGeneral は "config_general" オブジェクトに対応します。
type ConfigGeneral struct {
	ModelConfig                 ModelConfig `json:"model_config"`
	TotalEvaluationTimeSecondes string      `json:"total_evaluation_time_secondes"`
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
	TotalEvaluationTime  float64              `dynamodbav:"total_evaluation_time"`
}
