package lsp

// LSP protocol constants.
const (
	JSONRPCVersion = "2.0"

	SeverityError   = 1
	SeverityWarning = 2
	SeverityInfo    = 3
	SeverityHint    = 4

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
)

// Completion trigger kinds.
const (
	CompletionTriggerInvoked                  = 1
	CompletionTriggerCharacter                = 2
	CompletionTriggerForIncompleteCompletions = 3
)

// Completion item kinds.
const (
	CompletionItemKindProperty = 10
	CompletionItemKindKeyword  = 14
	CompletionItemKindUnit     = 11
	CompletionItemKindValue    = 12
	CompletionItemKindFunction = 3
	CompletionItemKindColor    = 16
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
