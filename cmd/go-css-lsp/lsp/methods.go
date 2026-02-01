// Package lsp implements LSP message types and handlers for
// CSS.
package lsp

import (
	"encoding/json"
	"errors"
	"log/slog"
	"strconv"
)

var filesOpenedByEditor = make(map[string]string)

// ID represents a JSON-RPC request ID (string or number).
type ID int

func (id *ID) UnmarshalJSON(data []byte) error {
	length := len(data)
	if data[0] == '"' && data[length-1] == '"' {
		data = data[1 : length-1]
	}

	number, err := strconv.Atoi(string(data))
	if err != nil {
		return errors.New(
			"'ID' expected either a string or an integer",
		)
	}

	*id = ID(number)
	return nil
}

func (id *ID) MarshalJSON() ([]byte, error) {
	val := strconv.Itoa(int(*id))
	return []byte(val), nil
}

// RequestMessage represents a JSON-RPC request.
type RequestMessage[T any] struct {
	JsonRpc string `json:"jsonrpc"`
	Id      ID     `json:"id"`
	Method  string `json:"method"`
	Params  T      `json:"params"`
}

// ResponseMessage represents a JSON-RPC response.
type ResponseMessage[T any] struct {
	JsonRpc string         `json:"jsonrpc"`
	Id      ID             `json:"id"`
	Result  T              `json:"result"`
	Error   *ResponseError `json:"error"`
}

// ResponseError represents a JSON-RPC error.
type ResponseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// NotificationMessage represents a JSON-RPC notification.
type NotificationMessage[T any] struct {
	JsonRpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  T      `json:"params"`
}

// InitializeParams holds parameters for initialize request.
type InitializeParams struct {
	ProcessId    int            `json:"processId"`
	Capabilities map[string]any `json:"capabilities"`
	ClientInfo   struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"clientInfo"`
	Locale           string `json:"locale"`
	RootUri          string `json:"rootUri"`
	Trace            any    `json:"trace"`
	WorkspaceFolders any    `json:"workspaceFolders"`
}

// ServerCapabilities describes this server's capabilities.
type ServerCapabilities struct {
	TextDocumentSync   int                `json:"textDocumentSync"`
	HoverProvider      bool               `json:"hoverProvider"`
	CompletionProvider *CompletionOptions `json:"completionProvider,omitempty"`
}

// CompletionOptions describes completion provider capabilities.
type CompletionOptions struct {
	TriggerCharacters []string `json:"triggerCharacters,omitempty"`
	ResolveProvider   bool     `json:"resolveProvider"`
}

// InitializeResult is the response to initialize.
type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`
	ServerInfo   struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"serverInfo"`
}

// PublishDiagnosticsParams holds diagnostic publishing params.
type PublishDiagnosticsParams struct {
	Uri         string       `json:"uri"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

// Diagnostic represents a diagnostic message.
type Diagnostic struct {
	Range    Range  `json:"range"`
	Message  string `json:"message"`
	Severity int    `json:"severity"`
}

// Position represents a position in a text document.
type Position struct {
	Line      uint `json:"line"`
	Character uint `json:"character"`
}

// Range represents a range in a text document.
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// TextDocumentItem represents a text document.
type TextDocumentItem struct {
	Uri        string `json:"uri"`
	Version    int    `json:"version"`
	LanguageId string `json:"languageId"`
	Text       string `json:"text"`
}

// TextDocumentIdentifier identifies a text document.
type TextDocumentIdentifier struct {
	Uri string `json:"uri"`
}

// TextDocumentPositionParams combines document + position.
type TextDocumentPositionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

// MarkupContent represents markup content.
type MarkupContent struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

// CompletionItem represents a completion suggestion.
type CompletionItem struct {
	Label         string         `json:"label"`
	Kind          int            `json:"kind,omitempty"`
	Detail        string         `json:"detail,omitempty"`
	Documentation *MarkupContent `json:"documentation,omitempty"`
	InsertText    string         `json:"insertText,omitempty"`
}

// CompletionList represents a list of completion items.
type CompletionList struct {
	IsIncomplete bool             `json:"isIncomplete"`
	Items        []CompletionItem `json:"items"`
}

// ProcessInitializeRequest handles the initialize request.
func ProcessInitializeRequest(
	data []byte,
	lspName, lspVersion string,
) (response []byte, root string) {
	req := RequestMessage[InitializeParams]{}

	err := json.Unmarshal(data, &req)
	if err != nil {
		msg := "error unmarshalling 'initialize': " + err.Error()
		slog.Error(msg,
			slog.Group("details",
				slog.String("received_req", string(data)),
			),
		)
		panic(msg)
	}

	res := ResponseMessage[InitializeResult]{
		JsonRpc: JSONRPCVersion,
		Id:      req.Id,
		Result: InitializeResult{
			Capabilities: ServerCapabilities{
				TextDocumentSync: TextDocumentSyncFull,
				HoverProvider:    true,
				CompletionProvider: &CompletionOptions{
					TriggerCharacters: []string{
						":", "@", ".", "#", "-", " ",
					},
					ResolveProvider: false,
				},
			},
		},
	}

	res.Result.ServerInfo.Name = lspName
	res.Result.ServerInfo.Version = lspVersion

	response, err = json.Marshal(res)
	if err != nil {
		msg := "error marshalling 'initialize': " + err.Error()
		slog.Error(msg)
		panic(msg)
	}

	return response, req.Params.RootUri
}

// ProcessInitializedNotification handles the initialized
// notification.
func ProcessInitializedNotification(data []byte) {
	slog.Info(
		"Received 'initialized' notification",
		slog.String("data", string(data)),
	)
}

// ProcessShutdownRequest handles shutdown.
func ProcessShutdownRequest(
	jsonVersion string,
	requestId ID,
) []byte {
	response := ResponseMessage[any]{
		JsonRpc: jsonVersion,
		Id:      requestId,
		Result:  nil,
		Error:   nil,
	}

	responseText, err := json.Marshal(response)
	if err != nil {
		msg := "Error marshalling shutdown response: " + err.Error()
		slog.Error(msg)
		panic(msg)
	}

	return responseText
}

// ProcessIllegalRequestAfterShutdown returns an error for
// post-shutdown requests.
func ProcessIllegalRequestAfterShutdown(
	jsonVersion string,
	requestId ID,
) []byte {
	response := ResponseMessage[any]{
		JsonRpc: jsonVersion,
		Id:      requestId,
		Result:  nil,
		Error: &ResponseError{
			Code:    ErrorInvalidRequest,
			Message: "illegal request while server shutting down",
		},
	}

	responseText, err := json.Marshal(response)
	if err != nil {
		msg := "Error marshalling error response: " + err.Error()
		slog.Error(msg)
		panic(msg)
	}

	return responseText
}

// DidOpenTextDocumentParams for textDocument/didOpen.
type DidOpenTextDocumentParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

// ProcessDidOpenTextDocumentNotification handles didOpen.
func ProcessDidOpenTextDocumentNotification(
	data []byte,
) (fileURI string, fileContent []byte) {
	request := RequestMessage[DidOpenTextDocumentParams]{}

	err := json.Unmarshal(data, &request)
	if err != nil {
		msg := "error unmarshalling 'textDocument/didOpen': " + err.Error()
		slog.Error(msg,
			slog.Group("details",
				slog.String("received_req", string(data)),
			),
		)
		panic(msg)
	}

	documentURI := request.Params.TextDocument.Uri
	documentContent := request.Params.TextDocument.Text
	filesOpenedByEditor[documentURI] = documentContent

	return documentURI, []byte(documentContent)
}

// TextDocumentContentChangeEvent represents a content change.
type TextDocumentContentChangeEvent struct {
	Range       Range  `json:"range"`
	RangeLength uint   `json:"rangeLength"`
	Text        string `json:"text"`
}

// DidChangeTextDocumentParams for textDocument/didChange.
type DidChangeTextDocumentParams struct {
	TextDocument   TextDocumentItem                 `json:"textDocument"`
	ContentChanges []TextDocumentContentChangeEvent `json:"contentChanges"`
}

// ProcessDidChangeTextDocumentNotification handles didChange.
func ProcessDidChangeTextDocumentNotification(
	data []byte,
) (fileURI string, fileContent []byte) {
	var request RequestMessage[DidChangeTextDocumentParams]

	err := json.Unmarshal(data, &request)
	if err != nil {
		msg := "error unmarshalling 'textDocument/didChange': " + err.Error()
		slog.Error(msg,
			slog.Group("details",
				slog.String("received_req", string(data)),
			),
		)
		panic(msg)
	}

	documentChanges := request.Params.ContentChanges
	if len(documentChanges) > 1 {
		msg := "server doesn't handle incremental changes yet"
		slog.Error(msg,
			slog.Group("details",
				slog.String("received_req", string(data)),
			),
		)
		panic(msg)
	}

	if len(documentChanges) == 0 {
		slog.Warn("'documentChanges' field is empty")
		return "", nil
	}

	documentContent := documentChanges[0].Text
	documentURI := request.Params.TextDocument.Uri
	filesOpenedByEditor[documentURI] = documentContent

	return documentURI, []byte(documentContent)
}

// DidCloseTextDocumentParams for textDocument/didClose.
type DidCloseTextDocumentParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

// ProcessDidCloseTextDocumentNotification handles didClose.
func ProcessDidCloseTextDocumentNotification(
	data []byte,
) (fileURI string, fileContent []byte) {
	var request RequestMessage[DidCloseTextDocumentParams]

	err := json.Unmarshal(data, &request)
	if err != nil {
		msg := "error unmarshalling 'textDocument/didClose': " + err.Error()
		slog.Error(msg,
			slog.Group("details",
				slog.String("received_req", string(data)),
			),
		)
		panic(msg)
	}

	documentPath := request.Params.TextDocument.Uri
	documentContent := request.Params.TextDocument.Text
	delete(filesOpenedByEditor, documentPath)

	return documentPath, []byte(documentContent)
}

// HoverParams for textDocument/hover.
type HoverParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

// HoverResult for textDocument/hover response.
type HoverResult struct {
	Contents MarkupContent `json:"contents"`
}

// CompletionParams for textDocument/completion.
type CompletionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
	Context      *CompletionContext     `json:"context,omitempty"`
}

// CompletionContext provides completion trigger info.
type CompletionContext struct {
	TriggerKind      int    `json:"triggerKind"`
	TriggerCharacter string `json:"triggerCharacter,omitempty"`
}
