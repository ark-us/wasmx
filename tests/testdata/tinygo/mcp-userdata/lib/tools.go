package lib

import (
	"encoding/json"
	"fmt"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

// ExecuteTool handles tool execution requests
func ExecuteTool(req ExecuteToolRequest) []byte {
	var response ExecuteToolResponse

	switch req.ToolName {
	case "set_favorite_color":
		response = setFavoriteColor(req.UserID, req.Arguments)
	case "get_favorite_color":
		response = getFavoriteColor(req.UserID, req.Arguments)
	case "list_items":
		response = listItems(req.UserID, req.Arguments)
	default:
		response = ExecuteToolResponse{
			Content: []ContentItem{{Type: "text", Text: "Unknown tool: " + req.ToolName}},
			IsError: true,
		}
	}

	data, _ := json.Marshal(response)
	return data
}

// setFavoriteColor sets the user's favorite color
func setFavoriteColor(userID string, arguments map[string]interface{}) ExecuteToolResponse {
	fmt.Println("=== setFavoriteColor called ===")
	fmt.Println("UserID:", userID)
	fmt.Println("Arguments:", arguments)

	if userID == "" {
		return ExecuteToolResponse{
			Content: []ContentItem{{Type: "text", Text: "Authentication required: no user_id provided"}},
			IsError: true,
		}
	}

	color, ok := arguments["color"].(string)
	if !ok {
		fmt.Println("Color argument not found or invalid")
		return ExecuteToolResponse{
			Content: []ContentItem{{Type: "text", Text: "Missing or invalid 'color' argument"}},
			IsError: true,
		}
	}

	keyStr := STORAGE_FAVORITE_COLOR_PREFIX + userID
	fmt.Println("Storage key:", keyStr)
	fmt.Println("Color value:", color)

	key := []byte(keyStr)
	wasmx.StorageStore(key, []byte(color))

	// Verify it was stored
	verifyData := wasmx.StorageLoad(key)
	fmt.Println("Verification - stored data:", string(verifyData))

	LoggerInfo("Favorite color set", []string{"user_id", userID, "color", color})

	return ExecuteToolResponse{
		Content: []ContentItem{{
			Type: "text",
			Text: fmt.Sprintf("Successfully set favorite color to %s for user %s", color, userID),
		}},
		IsError: false,
	}
}

// getFavoriteColor gets the user's favorite color
func getFavoriteColor(userID string, arguments map[string]interface{}) ExecuteToolResponse {
	fmt.Println("=== getFavoriteColor called ===")
	fmt.Println("UserID:", userID)

	if userID == "" {
		return ExecuteToolResponse{
			Content: []ContentItem{{Type: "text", Text: "Authentication required: no user_id provided"}},
			IsError: true,
		}
	}

	keyStr := STORAGE_FAVORITE_COLOR_PREFIX + userID
	fmt.Println("Storage key:", keyStr)

	key := []byte(keyStr)
	data := wasmx.StorageLoad(key)
	fmt.Println("Loaded data length:", len(data))
	fmt.Println("Loaded data:", string(data))

	if len(data) == 0 {
		fmt.Println("No data found in storage")
		return ExecuteToolResponse{
			Content: []ContentItem{{
				Type: "text",
				Text: fmt.Sprintf("No favorite color set for user %s", userID),
			}},
			IsError: false,
		}
	}

	color := string(data)
	fmt.Println("Returning color:", color)

	return ExecuteToolResponse{
		Content: []ContentItem{{
			Type: "text",
			Text: fmt.Sprintf("User %s's favorite color is %s", userID, color),
		}},
		IsError: false,
	}
}

// listItems lists all items for a user
func listItems(userID string, arguments map[string]interface{}) ExecuteToolResponse {
	if userID == "" {
		return ExecuteToolResponse{
			Content: []ContentItem{{Type: "text", Text: "Authentication required: no user_id provided"}},
			IsError: true,
		}
	}

	key := []byte(STORAGE_ITEMS_PREFIX + userID)
	data := wasmx.StorageLoad(key)

	// If no items exist, initialize with default items
	if len(data) == 0 {
		defaultItems := []Item{
			{Name: "water", Description: "H2O"},
			{Name: "salt", Description: "NaCl"},
			{Name: "pepper", Description: "Ground pepper"},
		}

		// Store the default items
		itemsData, _ := json.Marshal(defaultItems)
		wasmx.StorageStore(key, itemsData)

		// Format default items list
		text := "water, salt, pepper"
		return ExecuteToolResponse{
			Content: []ContentItem{{
				Type: "text",
				Text: text,
			}},
			IsError: false,
		}
	}

	var items []Item
	if err := json.Unmarshal(data, &items); err != nil {
		return ExecuteToolResponse{
			Content: []ContentItem{{
				Type: "text",
				Text: fmt.Sprintf("Error parsing items: %s", err.Error()),
			}},
			IsError: true,
		}
	}

	// Format items list as simple comma-separated names
	text := ""
	for i, item := range items {
		if i > 0 {
			text += ", "
		}
		text += item.Name
	}

	return ExecuteToolResponse{
		Content: []ContentItem{{
			Type: "text",
			Text: text,
		}},
		IsError: false,
	}
}
