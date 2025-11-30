package lib

import (
	"encoding/json"
	"fmt"

	wasmxcore "github.com/loredanacirstea/wasmx-env-core/lib"
)

// ExecuteTool handles tool execution requests
func ExecuteTool(req ExecuteToolRequest) []byte {

	var response ExecuteToolResponse

	switch req.ToolName {
	case "execute_py":
		response = executePy(req.Arguments)
	case "execute_cli":
		response = executeCli(req.Arguments)
	default:
		response = ExecuteToolResponse{
			Content: []ContentItem{{Type: "text", Text: "Unknown tool: " + req.ToolName}},
			IsError: true,
		}
	}

	data, _ := json.Marshal(response)
	return data
}

// executePy executes the Python hello.py script
func executePy(arguments map[string]interface{}) ExecuteToolResponse {
	scriptPath := "/Users/user/dev/blockchain/wasmx/tests/testdata/tinygo/mcp/hello.py"

	result, err := wasmxcore.ExecuteCliCommand("python3", []string{scriptPath}, "")
	if err != nil {
		return ExecuteToolResponse{
			Content: []ContentItem{{
				Type: "text",
				Text: fmt.Sprintf("Error executing Python script: %s", err.Error()),
			}},
			IsError: true,
		}
	}

	return ExecuteToolResponse{
		Content: []ContentItem{{
			Type: "text",
			Text: result.Stdout,
		}},
		IsError: false,
	}
}

// executeCli executes a CLI command
func executeCli(arguments map[string]interface{}) ExecuteToolResponse {
	// Extract command and args from arguments
	command, ok := arguments["command"].(string)
	if !ok {
		return ExecuteToolResponse{
			Content: []ContentItem{{Type: "text", Text: "Missing or invalid 'command' argument"}},
			IsError: true,
		}
	}

	// Extract args array if provided
	var args []string
	if argsInterface, ok := arguments["args"]; ok {
		if argsArray, ok := argsInterface.([]interface{}); ok {
			for _, arg := range argsArray {
				if argStr, ok := arg.(string); ok {
					args = append(args, argStr)
				}
			}
		}
	}

	// Extract stdin if provided
	stdin := ""
	if stdinInterface, ok := arguments["stdin"]; ok {
		if stdinStr, ok := stdinInterface.(string); ok {
			stdin = stdinStr
		}
	}

	result, err := wasmxcore.ExecuteCliCommand(command, args, stdin)
	if err != nil {
		return ExecuteToolResponse{
			Content: []ContentItem{{
				Type: "text",
				Text: fmt.Sprintf("Error executing command: %s", err.Error()),
			}},
			IsError: true,
		}
	}

	// Include both stdout and stderr in response
	output := result.Stdout
	if result.Stderr != "" {
		output += "\n--- stderr ---\n" + result.Stderr
	}

	return ExecuteToolResponse{
		Content: []ContentItem{{
			Type: "text",
			Text: output,
		}},
		IsError: result.ExitCode != 0,
	}
}
