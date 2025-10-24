package listadaccounts

// ToolInput defines the parameters accepted by the MCP tool.
type ToolInput struct {
	AccountIDs   []string `json:"account_ids,omitempty"`
	AccountNames []string `json:"account_names,omitempty"`
}
