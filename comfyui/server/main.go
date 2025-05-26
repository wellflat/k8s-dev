package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"time"
	"github.com/gin-gonic/gin"
)

//go:embed SD3.5M_example_workflow_api.json
var workflowContent []byte

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
	// Use embedded content instead of reading from file
	/*
	content, err := os.ReadFile(g.WorkflowPath)
	if err != nil {
		return nil, fmt.Errorf("workflow file not found: %s", g.WorkflowPath)
	}*/
	var workflow map[string]interface{}
	if err := json.Unmarshal(workflowContent, &workflow); err != nil {
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

func (g *ComfyUIPromptGenerator) QueuePrompt() (map[string]interface{}, error) {
	data := map[string]interface{}{
		"prompt": g.Prompt,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal prompt to JSON: %w", err)
	}

	url := fmt.Sprintf("http://%s/api/prompt", g.ServerAddress)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send prompt to server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var body map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			return nil, fmt.Errorf("failed to queue prompt: server returned status %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("failed to queue prompt: server returned status %d, response: %v", resp.StatusCode, body)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode server response: %w", err)
	}
	return response, nil
}

func (g *ComfyUIPromptGenerator) GenerateAndQueue(clipText string) (map[string]interface{}, error) {
	g.SetCLIPText(clipText, "")
	g.SetRandomSeed("")
	return g.QueuePrompt()
}

func main() {
	customWorkflowPath := "./SD3.5M_example_workflow_api.json"
	customGenerator, err := NewComfyUIPromptGenerator("", customWorkflowPath)
	if err != nil {
		fmt.Println("Error creating generator:", err)
		return
	}

	engine := gin.Default()
	engine.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "hello comfy",
		})
	})

	engine.GET("/system_stats", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "System stats endpoint",
		})
		// Here you can add logic to return ComfyUI system stats
	})

	engine.POST("/prompt", func(c *gin.Context) {
		var requestBody map[string]interface{}
		if err := c.ShouldBindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
		res, err := customGenerator.GenerateAndQueue(requestBody["prompt"].(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to queue prompt"})
			return
		}
		fmt.Println("Received prompt:", requestBody["prompt"])
		fmt.Println("Response from ComfyUI:", res["prompt_id"])
		c.JSON(http.StatusOK, gin.H{"status": "success", "prompt_id": res["prompt_id"]})
	})
	engine.Run(":3000")
}