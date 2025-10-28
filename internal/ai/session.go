package ai

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// ChatSession provides a stateful interface for multi-turn conversations with AI providers.
// It manages message history, function definitions, and session-specific configuration.
type ChatSession struct {
	session  *types.ChatSession
	provider Provider
}

// NewChatSession creates a new chat session with the given provider and configuration.
func NewChatSession(provider Provider, config types.SessionConfig) *ChatSession {
	return &ChatSession{
		session: &types.ChatSession{
			ID:        uuid.New().String(),
			Messages:  make([]types.Message, 0),
			Functions: make([]types.FunctionDefinition, 0),
			Config:    config,
			Metadata:  make(map[string]string),
		},
		provider: provider,
	}
}

// ID returns the unique session identifier.
func (s *ChatSession) ID() string {
	return s.session.ID
}

// AddMessage appends a message to the conversation history.
// Valid roles: "system", "user", "assistant", "function".
func (s *ChatSession) AddMessage(role, content string) error {
	if role == "" {
		return fmt.Errorf("message role cannot be empty")
	}

	if content == "" {
		return fmt.Errorf("message content cannot be empty")
	}

	// Validate role
	validRoles := map[string]bool{
		"system":    true,
		"user":      true,
		"assistant": true,
		"function":  true,
	}

	if !validRoles[role] {
		return fmt.Errorf("invalid message role: %s (valid: system, user, assistant, function)", role)
	}

	message := types.Message{
		Role:      role,
		Content:   content,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	s.session.Messages = append(s.session.Messages, message)
	return nil
}

// AddMessageWithFunction appends an assistant message with a function call.
func (s *ChatSession) AddMessageWithFunction(functionCall *types.FunctionCall) error {
	if functionCall == nil {
		return fmt.Errorf("function call cannot be nil")
	}

	if functionCall.Name == "" {
		return fmt.Errorf("function name cannot be empty")
	}

	message := types.Message{
		Role:         "assistant",
		Content:      "",
		FunctionCall: functionCall,
		Timestamp:    time.Now().Format(time.RFC3339),
	}

	s.session.Messages = append(s.session.Messages, message)
	return nil
}

// AddFunctionResult appends a function result message.
func (s *ChatSession) AddFunctionResult(functionName string, result interface{}) error {
	if functionName == "" {
		return fmt.Errorf("function name cannot be empty")
	}

	message := types.Message{
		Role:           "function",
		Content:        functionName,
		FunctionResult: result,
		Timestamp:      time.Now().Format(time.RFC3339),
	}

	s.session.Messages = append(s.session.Messages, message)
	return nil
}

// SetFunctions registers available tools/functions for the session.
// This replaces any previously set functions.
func (s *ChatSession) SetFunctions(functions []types.FunctionDefinition) error {
	if functions == nil {
		return fmt.Errorf("functions cannot be nil (use empty slice if no functions)")
	}

	// Validate function definitions
	for i, fn := range functions {
		if fn.Name == "" {
			return fmt.Errorf("function %d: name cannot be empty", i)
		}
		if fn.Description == "" {
			return fmt.Errorf("function %q: description cannot be empty", fn.Name)
		}
		if fn.Parameters == nil {
			return fmt.Errorf("function %q: parameters cannot be nil", fn.Name)
		}
	}

	s.session.Functions = functions
	return nil
}

// Send sends the current conversation to the AI provider and returns the response.
// The response is automatically added to the message history.
func (s *ChatSession) Send(ctx context.Context) (string, error) {
	if len(s.session.Messages) == 0 {
		return "", fmt.Errorf("cannot send empty conversation")
	}

	// Build prompt from message history
	prompt := s.buildPrompt()

	// Call provider
	response, err := s.provider.AnalyzeWithContext(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("provider call failed: %w", err)
	}

	// Add response to history
	if err := s.AddMessage("assistant", response); err != nil {
		return "", fmt.Errorf("failed to add response to history: %w", err)
	}

	return response, nil
}

// Reset clears the message history but preserves functions and metadata.
func (s *ChatSession) Reset() {
	s.session.Messages = make([]types.Message, 0)
}

// GetMessages returns a copy of the message history.
func (s *ChatSession) GetMessages() []types.Message {
	messages := make([]types.Message, len(s.session.Messages))
	copy(messages, s.session.Messages)
	return messages
}

// GetMetadata returns the session metadata.
func (s *ChatSession) GetMetadata() map[string]string {
	return s.session.Metadata
}

// SetMetadata sets a metadata key-value pair.
func (s *ChatSession) SetMetadata(key, value string) {
	s.session.Metadata[key] = value
}

// buildPrompt constructs a prompt from the message history.
// This is a simple implementation that concatenates messages.
// Provider-specific implementations may override this behavior.
func (s *ChatSession) buildPrompt() string {
	var prompt string

	for _, msg := range s.session.Messages {
		switch msg.Role {
		case "system":
			prompt += fmt.Sprintf("System: %s\n\n", msg.Content)
		case "user":
			prompt += fmt.Sprintf("User: %s\n\n", msg.Content)
		case "assistant":
			if msg.FunctionCall != nil {
				prompt += fmt.Sprintf("Assistant: [Function Call: %s]\n\n", msg.FunctionCall.Name)
			} else {
				prompt += fmt.Sprintf("Assistant: %s\n\n", msg.Content)
			}
		case "function":
			prompt += fmt.Sprintf("Function Result [%s]: %v\n\n", msg.Content, msg.FunctionResult)
		}
	}

	return prompt
}
