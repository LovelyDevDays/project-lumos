package hybrid

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the main configuration structure
type Config struct {
	Search   SearchConfig   `yaml:"search"`
	Database DatabaseConfig `yaml:"database"`
	API      APIConfig      `yaml:"api"`
}

// SearchConfig holds search-related configuration
type SearchConfig struct {
	EmbeddingWeight float32 `yaml:"embedding_weight"`
	BM42Weight      float32 `yaml:"bm42_weight"`
	MaxResults      int     `yaml:"max_results"`
	MinScore        float32 `yaml:"min_score"`
	OverlapBonus    float32 `yaml:"overlap_bonus"`    // Bonus multiplier for documents in multiple searches
	OverlapPriority bool    `yaml:"overlap_priority"` // Enable overlap prioritization
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	QdrantHost        string `yaml:"qdrant_host"`
	QdrantPort        int    `yaml:"qdrant_port"`
	CollectionName    string `yaml:"collection_name"`
	TitleCollection   string `yaml:"title_collection"`
	ContentCollection string `yaml:"content_collection"`
	BM42Collection    string `yaml:"bm42_collection"`
}

// APIConfig holds API server configuration
type APIConfig struct {
	EmbeddingServer string `yaml:"embedding_server"`
	Timeout         int    `yaml:"timeout"`
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(filepath string) (*Config, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Set defaults if not specified
	config.setDefaults()

	return &config, nil
}

// setDefaults sets default values for unspecified configuration
func (c *Config) setDefaults() {
	if c.Search.EmbeddingWeight == 0 && c.Search.BM42Weight == 0 {
		c.Search.EmbeddingWeight = 0.5
		c.Search.BM42Weight = 0.5
	}

	if c.Search.MaxResults == 0 {
		c.Search.MaxResults = 10
	}

	if c.Database.QdrantHost == "" {
		c.Database.QdrantHost = "localhost"
	}

	if c.Database.QdrantPort == 0 {
		c.Database.QdrantPort = 6334
	}

	if c.Database.CollectionName == "" {
		c.Database.CollectionName = "default_collection"
	}

	if c.API.EmbeddingServer == "" {
		c.API.EmbeddingServer = "http://localhost:8080/v1"
	}

	if c.API.Timeout == 0 {
		c.API.Timeout = 30
	}
}

// SaveConfig saves configuration to a YAML file
func SaveConfig(config *Config, filepath string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetDefaultConfig returns a default configuration
func GetDefaultConfig() *Config {
	config := &Config{
		Search: SearchConfig{
			EmbeddingWeight: 0.5,
			BM42Weight:      0.5,
			MaxResults:      10,
			MinScore:        0.1,
			OverlapBonus:    2.0,
			OverlapPriority: true,
		},
		Database: DatabaseConfig{
			QdrantHost:        "localhost",
			QdrantPort:        6334,
			CollectionName:    "jira_issues",
			TitleCollection:   "jira_titles",
			ContentCollection: "jira_contents",
			BM42Collection:    "jira_issues_bm42",
		},
		API: APIConfig{
			EmbeddingServer: "http://localhost:8080/v1",
			Timeout:         30,
		},
	}

	return config
}
