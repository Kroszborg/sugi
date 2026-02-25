package ai

import (
	"context"
	"fmt"
	"os"
)

// Generator wraps the Groq API client.
type Generator struct {
	groq *GroqClient
}

// NewGenerator creates a Generator using the provided Groq API key.
// If apiKey is empty it falls back to the GROQ_API_KEY environment variable.
func NewGenerator(apiKey, model string) *Generator {
	if apiKey == "" {
		apiKey = os.Getenv("GROQ_API_KEY")
	}
	g := &Generator{}
	if apiKey != "" {
		g.groq = NewGroqClient(apiKey, model)
	}
	return g
}

// Generate streams a response for the given prompt via Groq.
// onChunk is called with each streamed token.
func (g *Generator) Generate(ctx context.Context, prompt string, onChunk func(string)) error {
	if g.groq == nil || g.groq.apiKey == "" {
		return fmt.Errorf("Groq API key not set — add groq_api_key to ~/.config/sugi/config.json or set GROQ_API_KEY env var")
	}
	return g.groq.Generate(ctx, prompt, onChunk)
}

// Available reports whether a Groq API key is configured.
func (g *Generator) Available() bool {
	return g.groq != nil && g.groq.apiKey != ""
}

// BackendName returns the active backend description.
func (g *Generator) BackendName() string {
	if g.groq != nil && g.groq.apiKey != "" {
		return "Groq (" + g.groq.model + ")"
	}
	return "none — set GROQ_API_KEY"
}
