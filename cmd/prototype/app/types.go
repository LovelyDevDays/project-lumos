package app

type Embedding struct {
	Payload map[string]any `json:"payload"`
	Vectors []float32      `json:"vectors"`
}
