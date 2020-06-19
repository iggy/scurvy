// JSONRPC Marshal/Unmarshal data sctructures

//lint:file-ignore U1000 Ignore all unused code, these interfaces are defined elsewhere

package main

// JSONRPCRequest - A fairly generic request sent to us
type JSONRPCRequest struct {
	ID      int    `json:"id"`
	Method  string `json:"method"`
	JSONRPC string `json:"jsonrpc"`
}

// JSONRPCRequestParamsGUISN - A more specific request sent to us
type JSONRPCRequestParamsGUISN struct {
	JSONRPCRequest
	Params struct {
		Title   string `json:"title"`
		Message string `json:"message"`
		Image   string `json:"image"`
	} `json:"params"`
}

// JSONRPCGenericResponse - A generic response
type JSONRPCGenericResponse struct {
	ID      int    `json:"id"`
	JSONRPC string `json:"jsonrpc"`
	Result  string `json:"result"`
}

// JSONRPCVersionResponse - response for JSONRPC.Version method
type JSONRPCVersionResponse struct {
	ID      int    `json:"id"`
	JSONRPC string `json:"jsonrpc"`
	Result  struct {
		Version struct {
			Major int `json:"major"`
			Minor int `json:"minor"`
			Patch int `json:"patch"`
		} `json:"version"`
	} `json:"result"`
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
