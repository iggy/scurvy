// JSONRPC Marshal/Unmarshal data sctructures

package main

// JSONRPCRequest - A request sent to us
type JSONRPCRequest struct {
	ID      int    `json:"id"`
	Method  string `json:"method"`
	JSONRPC string `json:"jsonrpc"`
}

// JSONRPCGenericResponse - A generic response
type JSONRPCGenericResponse struct {
	ID      int    `json:"id"`
	JSONRPC string `json:"jsonrpc"`
	Result  string `json:"result"`
}

// JSONRPCVersion - JSONRPC.Version response
type JSONRPCVersion struct {
	Major int `json:"major"`
	Minor int `json:"minor"`
	Patch int `json:"patch"`
}

// JSONRPCVersionResult - part of JSONRPCVersionReply
type JSONRPCVersionResult struct {
	Version JSONRPCVersion `json:"version"`
}

// JSONRPCVersionResponse - response for JSONRPC.Version method
type JSONRPCVersionResponse struct {
	ID      int                  `json:"id"`
	JSONRPC string               `json:"jsonrpc"`
	Result  JSONRPCVersionResult `json:"result"`
}

// GUIShowNotificationResponse - GUI.ShowNotification response - uses JSONRPCGenericResponse
// type GUIShowNotificationResponse struct {
// }
// sabnzbd JSON Marshal/Unmarshal data sctructures

// SABJSONRequest - A request sent to us
type SABJSONRequest struct {
	Message string `json:"message"`
	Version string `json:"version"`
	Type    string `json:"type"`
	Title   string `json:"title"`
}
