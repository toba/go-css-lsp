package lsp

// LSP protocol constants.
const (
	JSONRPCVersion = "2.0"

	TextDocumentSyncFull = 1

	ErrorInvalidRequest = -32600
)

// LSP method names.
const (
	MethodInitialize         = "initialize"
	MethodInitialized        = "initialized"
	MethodShutdown           = "shutdown"
	MethodExit               = "exit"
	MethodDidOpen            = "textDocument/didOpen"
	MethodDidChange          = "textDocument/didChange"
	MethodDidClose           = "textDocument/didClose"
	MethodHover              = "textDocument/hover"
	MethodCompletion         = "textDocument/completion"
	MethodPublishDiagnostics = "textDocument/publishDiagnostics"
	MethodDocumentColor      = "textDocument/documentColor"
	MethodColorPresentation  = "textDocument/colorPresentation"
	MethodDocumentSymbol     = "textDocument/documentSymbol"
	MethodDefinition         = "textDocument/definition"
	MethodReferences         = "textDocument/references"
	MethodCodeAction         = "textDocument/codeAction"
	MethodDocumentHighlight  = "textDocument/documentHighlight"
	MethodFoldingRange       = "textDocument/foldingRange"
	MethodDocumentLink       = "textDocument/documentLink"
	MethodFormatting         = "textDocument/formatting"
	MethodSelectionRange     = "textDocument/selectionRange"
	MethodPrepareRename      = "textDocument/prepareRename"
	MethodRename             = "textDocument/rename"
)

// Completion trigger kinds.
const (
	CompletionTriggerInvoked                  = 1
	CompletionTriggerCharacter                = 2
	CompletionTriggerForIncompleteCompletions = 3
)

// LSP header constants.
const (
	ContentLengthHeader = "Content-Length"
	HeaderDelimiter     = "\r\n\r\n"
	LineDelimiter       = "\r\n"
)

// File and logging constants.
const (
	DirPermissions  = 0750
	FilePermissions = 0600
	MaxLogFileSize  = 5_000_000
)
