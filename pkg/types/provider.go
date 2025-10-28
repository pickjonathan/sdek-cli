package types

// ProviderConfig defines configuration for an AI provider.
type ProviderConfig struct {
	// URL is the provider URL with scheme (e.g., "openai://api.openai.com")
	URL string `yaml:"url" json:"url" mapstructure:"url"`

	// APIKey is the authentication key (supports ${VAR} substitution)
	APIKey string `yaml:"api_key,omitempty" json:"api_key,omitempty" mapstructure:"api_key"`

	// Model is the model identifier (e.g., "gpt-4o", "gemma3:12b")
	Model string `yaml:"model" json:"model" mapstructure:"model"`

	// Endpoint is an optional custom endpoint override
	Endpoint string `yaml:"endpoint,omitempty" json:"endpoint,omitempty" mapstructure:"endpoint"`

	// Timeout is the request timeout in seconds (default 60)
	Timeout int `yaml:"timeout" json:"timeout" mapstructure:"timeout"`

	// MaxRetries is the maximum retry attempts (default 3)
	MaxRetries int `yaml:"max_retries" json:"max_retries" mapstructure:"max_retries"`

	// Temperature is the sampling temperature 0.0-2.0 (default 0.0)
	Temperature float64 `yaml:"temperature" json:"temperature" mapstructure:"temperature"`

	// MaxTokens is the maximum response tokens (default 4096)
	MaxTokens int `yaml:"max_tokens" json:"max_tokens" mapstructure:"max_tokens"`

	// Extra contains provider-specific settings
	Extra map[string]string `yaml:"extra,omitempty" json:"extra,omitempty" mapstructure:"extra"`
}

// ChatSession represents a multi-turn conversation with an AI provider.
type ChatSession struct {
	// ID is the unique session identifier (UUID)
	ID string `json:"id"`

	// Messages contains the conversation history
	Messages []Message `json:"messages"`

	// Functions contains available tools/functions
	Functions []FunctionDefinition `json:"functions,omitempty"`

	// Config contains session-specific configuration
	Config SessionConfig `json:"config"`

	// Metadata contains custom metadata
	Metadata map[string]string `json:"metadata,omitempty"`
}

// SessionConfig contains session-specific configuration.
type SessionConfig struct {
	// Temperature is the sampling temperature
	Temperature float64 `json:"temperature"`

	// MaxTokens is the maximum response tokens
	MaxTokens int `json:"max_tokens"`
}

// Message represents a single message in a conversation.
type Message struct {
	// Role is the message sender: "system", "user", "assistant", or "function"
	Role string `json:"role"`

	// Content is the message text
	Content string `json:"content"`

	// FunctionCall contains optional function call (if role = "assistant")
	FunctionCall *FunctionCall `json:"function_call,omitempty"`

	// FunctionResult contains optional function result (if role = "function")
	FunctionResult interface{} `json:"function_result,omitempty"`

	// Timestamp is when the message was created
	Timestamp string `json:"timestamp"`
}

// FunctionCall represents a request to execute a function.
type FunctionCall struct {
	// Name is the function identifier
	Name string `json:"name"`

	// Arguments contains the function arguments as JSON
	Arguments map[string]interface{} `json:"arguments"`
}

// FunctionDefinition defines a tool/function available to the AI.
type FunctionDefinition struct {
	// Name is the function identifier
	Name string `json:"name"`

	// Description is the natural language description for AI
	Description string `json:"description"`

	// Parameters is the expected input schema (JSON Schema)
	Parameters map[string]interface{} `json:"parameters"`
}

// DefaultProviderConfig returns a default provider configuration.
func DefaultProviderConfig() ProviderConfig {
	return ProviderConfig{
		Timeout:     60,
		MaxRetries:  3,
		Temperature: 0.0,
		MaxTokens:   4096,
		Extra:       make(map[string]string),
	}
}
