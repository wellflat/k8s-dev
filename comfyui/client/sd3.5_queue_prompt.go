package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type ComfyUIPromptGenerator struct {
	ServerAddress string
	WorkflowPath  string
	Prompt        map[string]interface{}
}

func NewComfyUIPromptGenerator(serverAddress string, workflowPath string) (*ComfyUIPromptGenerator, error) {
	if serverAddress == "" {
		serverAddress = "127.0.0.1:8188"
	}

	if workflowPath == "" {
		executablePath, err := os.Executable()
		if err != nil {
			return nil, fmt.Errorf("failed to get executable path: %w", err)
		}
		workflowPath = filepath.Join(filepath.Dir(executablePath), "workflow_api.json")
	}

	generator := &ComfyUIPromptGenerator{
		ServerAddress: serverAddress,
		WorkflowPath:  workflowPath,
	}

	prompt, err := generator.loadWorkflow()
	if err != nil {
		return nil, err
	}
	generator.Prompt = prompt

	return generator, nil
}

func (g *ComfyUIPromptGenerator) loadWorkflow() (map[string]interface{}, error) {
	content, err := os.ReadFile(g.WorkflowPath)
	if err != nil {
		return nil, fmt.Errorf("workflow file not found: %s", g.WorkflowPath)
	}

	var workflow map[string]interface{}
	if err := json.Unmarshal(content, &workflow); err != nil {
		return nil, fmt.Errorf("invalid JSON in workflow file: %s - %w", g.WorkflowPath, err)
	}

	return workflow, nil
}

func (g *ComfyUIPromptGenerator) SetCLIPText(text string, nodeID string) {
	if nodeID == "" {
		nodeID = "6"
	}
	if node, ok := g.Prompt[nodeID].(map[string]interface{}); ok {
		if inputs, ok := node["inputs"].(map[string]interface{}); ok {
			inputs["text"] = text
		}
	}
}

func (g *ComfyUIPromptGenerator) SetRandomSeed(nodeID string) {
	if nodeID == "" {
		nodeID = "294"
	}
	rand.Seed(time.Now().UnixNano())
	seed := rand.Intn(1_000_000) + 1
	if node, ok := g.Prompt[nodeID].(map[string]interface{}); ok {
		if inputs, ok := node["inputs"].(map[string]interface{}); ok {
			inputs["seed"] = seed
		}
	}
}

func (g *ComfyUIPromptGenerator) QueuePrompt() error {
	data := map[string]interface{}{
		"prompt": g.Prompt,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal prompt to JSON: %w", err)
	}

	url := fmt.Sprintf("http://%s/api/prompt", g.ServerAddress)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send prompt to server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var body map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			return fmt.Errorf("failed to queue prompt: server returned status %d", resp.StatusCode)
		}
		return fmt.Errorf("failed to queue prompt: server returned status %d, response: %v", resp.StatusCode, body)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to decode server response: %w", err)
	}
	fmt.Println(response)

	return nil
}

func (g *ComfyUIPromptGenerator) GenerateAndQueue(clipText string) error {
	g.SetCLIPText(clipText, "") // デフォルトのノードIDを使用
	g.SetRandomSeed("")        // デフォルトのノードIDを使用
	return g.QueuePrompt()
}

func main() {
	customWorkflowPath := "SD3.5M_example_workflow_api.json"
	customGenerator, err := NewComfyUIPromptGenerator("", customWorkflowPath)
	if err != nil {
		fmt.Println("Error creating generator:", err)
		return
	}

	prompt := "A beautiful girl in a fantasy world"
	err = customGenerator.GenerateAndQueue(prompt)
	if err != nil {
		fmt.Println("Error generating and queueing prompt:", err)
	}
}