// Package local_llm adapts Ollama (and LM Studio) as UCLAW LLM providers.
// Ollama: https://github.com/ollama/ollama
// LM Studio: https://github.com/lmstudio-ai/lmstudio-js
package local_llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Config holds connection settings for a local model server.
type Config struct {
	BaseURL string // e.g. http://localhost:11434
	Model   string // e.g. llama3
}

type generateReq struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type generateResp struct {
	Response string `json:"response"`
}

// Generate sends a prompt and returns the model response.
func (c *Config) Generate(prompt string) (string, error) {
	body, _ := json.Marshal(generateReq{Model: c.Model, Prompt: prompt, Stream: false})
	resp, err := http.Post(c.BaseURL+"/api/generate", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("local_llm: generate: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var gr generateResp
	if err := json.Unmarshal(raw, &gr); err != nil {
		return "", fmt.Errorf("local_llm: decode: %w (body: %s)", err, raw)
	}
	return gr.Response, nil
}

type embedReq struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type embedResp struct {
	Embedding []float64 `json:"embedding"`
}

// Embed returns an embedding vector for the given text.
func (c *Config) Embed(text string) ([]float64, error) {
	body, _ := json.Marshal(embedReq{Model: c.Model, Prompt: text})
	resp, err := http.Post(c.BaseURL+"/api/embeddings", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("local_llm: embed: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var er embedResp
	if err := json.Unmarshal(raw, &er); err != nil {
		return nil, fmt.Errorf("local_llm: embed decode: %w", err)
	}
	return er.Embedding, nil
}
