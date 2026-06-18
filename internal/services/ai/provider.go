package ai

import "context"

// AnalysisResult is the structured payload every AI provider is
// expected to produce.
type AnalysisResult struct {
	Summary      string `json:"summary"`
	Sentiment    string `json:"sentiment"`     // positive | negative | neutral
	Hook         string `json:"hook"`
	Tweet        string `json:"tweet"`
	ThreadOpener string `json:"thread_opener"`
}

// Provider is implemented by any OpenAI-compatible chat completions
// endpoint (OpenAI, OpenRouter, Together, Groq, local llama.cpp, ...).
type Provider interface {
	// Analyze sends `prompt` to the underlying model and returns the raw
	// assistant content. The provider is expected to return a JSON
	// document matching AnalysisResult.
	Analyze(ctx context.Context, prompt string) (string, error)
	// Name is a stable identifier (e.g. "openai-compatible").
	Name() string
}
