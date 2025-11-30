package lib

const MODULE_NAME = "mcp-execute"

// CallData represents the different operations this contract supports
type CallData struct {
	// Tool execution
	ExecuteTool *ExecuteToolRequest `json:"execute_tool,omitempty"`
}

// ExecuteToolRequest represents a request to execute a tool
type ExecuteToolRequest struct {
	ToolName  string                 `json:"tool_name"`
	Arguments map[string]interface{} `json:"arguments"`
	UserID    string                 `json:"user_id,omitempty"`
}

// ExecuteToolResponse represents the response from tool execution
type ExecuteToolResponse struct {
	Content []ContentItem `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

// ContentItem represents a single content item in the response
type ContentItem struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}
