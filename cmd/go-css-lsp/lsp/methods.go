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

// ServerSettings holds server-specific configuration from
// initializationOptions.
type ServerSettings struct {
	FormatMode           string `json:"formatMode"`
	PrintWidth           int    `json:"printWidth"`
	ExperimentalFeatures string `json:"experimentalFeatures"`
	DeprecatedFeatures   string `json:"deprecatedFeatures"`
	UnknownValues        string `json:"unknownValues"`
	StrictColorNames     bool   `json:"strictColorNames"`
}

// InitializeParams holds parameters for initialize request.
type InitializeParams struct {
	ProcessId    int            `json:"processId"`
	Capabilities map[string]any `json:"capabilities"`
	ClientInfo   struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"clientInfo"`
	Locale                string          `json:"locale"`
	RootUri               string          `json:"rootUri"`
	Trace                 any             `json:"trace"`
	WorkspaceFolders      any             `json:"workspaceFolders"`
	InitializationOptions *ServerSettings `json:"initializationOptions,omitempty"`
}

// ServerCapabilities describes this server's capabilities.
type ServerCapabilities struct {
	TextDocumentSync           int                  `json:"textDocumentSync"`
	HoverProvider              bool                 `json:"hoverProvider"`
	CompletionProvider         *CompletionOptions   `json:"completionProvider,omitempty"`
	ColorProvider              bool                 `json:"colorProvider,omitempty"`
	DocumentSymbolProvider     bool                 `json:"documentSymbolProvider,omitempty"`
	DefinitionProvider         bool                 `json:"definitionProvider,omitempty"`
	ReferencesProvider         bool                 `json:"referencesProvider,omitempty"`
	CodeActionProvider         *CodeActionOptions   `json:"codeActionProvider,omitempty"`
	DocumentHighlightProvider  bool                 `json:"documentHighlightProvider,omitempty"`
	FoldingRangeProvider       bool                 `json:"foldingRangeProvider,omitempty"`
	DocumentLinkProvider       *DocumentLinkOptions `json:"documentLinkProvider,omitempty"`
	DocumentFormattingProvider bool                 `json:"documentFormattingProvider,omitempty"`
	SelectionRangeProvider     bool                 `json:"selectionRangeProvider,omitempty"`
	RenameProvider             *RenameOptions       `json:"renameProvider,omitempty"`
}

// DocumentLinkOptions describes document link provider capabilities.
type DocumentLinkOptions struct {
	ResolveProvider bool `json:"resolveProvider,omitempty"`
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
	Tags          []int          `json:"tags,omitempty"`
	Deprecated    bool           `json:"deprecated,omitempty"`
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
) (response []byte, root string, settings *ServerSettings) {
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
				TextDocumentSync:       TextDocumentSyncFull,
				HoverProvider:          true,
				ColorProvider:          true,
				DocumentSymbolProvider: true,
				DefinitionProvider:     true,
				ReferencesProvider:     true,
				CodeActionProvider: &CodeActionOptions{
					CodeActionKinds: []string{
						"quickfix", "refactor", "source.fixAll",
					},
				},
				DocumentHighlightProvider:  true,
				FoldingRangeProvider:       true,
				DocumentLinkProvider:       &DocumentLinkOptions{},
				DocumentFormattingProvider: true,
				SelectionRangeProvider:     true,
				RenameProvider: &RenameOptions{
					PrepareProvider: true,
				},
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

	return response, req.Params.RootUri, req.Params.InitializationOptions
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
	Range    *Range        `json:"range,omitempty"`
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

// DocumentColorParams for textDocument/documentColor.
type DocumentColorParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// LSPColor represents a color in the LSP protocol.
type LSPColor struct {
	Red   float64 `json:"red"`
	Green float64 `json:"green"`
	Blue  float64 `json:"blue"`
	Alpha float64 `json:"alpha"`
}

// ColorInformation represents a color range in a document.
type ColorInformation struct {
	Range Range    `json:"range"`
	Color LSPColor `json:"color"`
}

// ColorPresentationParams for textDocument/colorPresentation.
type ColorPresentationParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Color        LSPColor               `json:"color"`
	Range        Range                  `json:"range"`
}

// ColorPresentation represents how a color is presented.
type ColorPresentation struct {
	Label    string    `json:"label"`
	TextEdit *TextEdit `json:"textEdit,omitempty"`
}

// TextEdit represents a text edit.
type TextEdit struct {
	Range   Range  `json:"range"`
	NewText string `json:"newText"`
}

// DocumentSymbolParams for textDocument/documentSymbol.
type DocumentSymbolParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// LSPDocumentSymbol represents a symbol in a document.
type LSPDocumentSymbol struct {
	Name           string              `json:"name"`
	Kind           int                 `json:"kind"`
	Range          Range               `json:"range"`
	SelectionRange Range               `json:"selectionRange"`
	Children       []LSPDocumentSymbol `json:"children,omitempty"`
}

// ReferenceParams for textDocument/references.
type ReferenceParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
	Context      ReferenceContext       `json:"context"`
}

// ReferenceContext for reference requests.
type ReferenceContext struct {
	IncludeDeclaration bool `json:"includeDeclaration"`
}

// LSPLocation represents a location in a document.
type LSPLocation struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

// LSPLocationLink represents a link between a source and
// target location, with explicit origin selection range.
type LSPLocationLink struct {
	OriginSelectionRange *Range `json:"originSelectionRange,omitempty"`
	TargetUri            string `json:"targetUri"`
	TargetRange          Range  `json:"targetRange"`
	TargetSelectionRange Range  `json:"targetSelectionRange"`
}

// CodeActionParams for textDocument/codeAction.
type CodeActionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Range        Range                  `json:"range"`
	Context      CodeActionContext      `json:"context"`
}

// CodeActionContext carries diagnostics for code actions.
type CodeActionContext struct {
	Diagnostics []Diagnostic `json:"diagnostics"`
	Only        []string     `json:"only,omitempty"`
}

// LSPCodeAction represents a code action response.
type LSPCodeAction struct {
	Title string         `json:"title"`
	Kind  string         `json:"kind,omitempty"`
	Edit  *WorkspaceEdit `json:"edit,omitempty"`
}

// WorkspaceEdit represents changes to workspace resources.
type WorkspaceEdit struct {
	Changes map[string][]TextEdit `json:"changes,omitempty"`
}

// LSPDocumentHighlight represents a highlighted range.
type LSPDocumentHighlight struct {
	Range Range `json:"range"`
	Kind  int   `json:"kind"`
}

// FoldingRangeParams for textDocument/foldingRange.
type FoldingRangeParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// LSPFoldingRange represents a foldable range.
type LSPFoldingRange struct {
	StartLine      int    `json:"startLine"`
	StartCharacter int    `json:"startCharacter,omitempty"`
	EndLine        int    `json:"endLine"`
	EndCharacter   int    `json:"endCharacter,omitempty"`
	Kind           string `json:"kind,omitempty"`
}

// DocumentLinkParams for textDocument/documentLink.
type DocumentLinkParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// LSPDocumentLink represents a link in a document.
type LSPDocumentLink struct {
	Range  Range  `json:"range"`
	Target string `json:"target"`
}

// DocumentFormattingParams for textDocument/formatting.
type DocumentFormattingParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Options      FormattingOptions      `json:"options"`
}

// FormattingOptions describes formatting preferences.
type FormattingOptions struct {
	TabSize      int  `json:"tabSize"`
	InsertSpaces bool `json:"insertSpaces"`
}

// CodeActionOptions describes code action provider capabilities.
type CodeActionOptions struct {
	CodeActionKinds []string `json:"codeActionKinds,omitempty"`
}

// RenameOptions for rename provider capability.
type RenameOptions struct {
	PrepareProvider bool `json:"prepareProvider"`
}

// SelectionRangeParams for textDocument/selectionRange.
type SelectionRangeParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Positions    []Position             `json:"positions"`
}

// LSPSelectionRange represents a selection range.
type LSPSelectionRange struct {
	Range  Range              `json:"range"`
	Parent *LSPSelectionRange `json:"parent,omitempty"`
}

// RenameParams for textDocument/rename.
type RenameParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
	NewName      string                 `json:"newName"`
}
