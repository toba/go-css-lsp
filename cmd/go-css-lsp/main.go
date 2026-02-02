// Command go-css-lsp provides a Language Server Protocol server
// for CSS.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"maps"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"sync"

	"github.com/toba/go-css-lsp/cmd/go-css-lsp/lsp"
	"github.com/toba/go-css-lsp/internal/css"
	"github.com/toba/go-css-lsp/internal/css/analyzer"
	"github.com/toba/go-css-lsp/internal/css/parser"
	"github.com/toba/go-css-lsp/internal/css/workspace"
)

// version is set by goreleaser at build time.
var version = "dev"

// workspaceStore holds the state for a workspace.
type workspaceStore struct {
	mu          sync.RWMutex
	RootPath    string
	RawFiles    map[string][]byte
	ParsedFiles map[string]*parser.Stylesheet
	VarIndex    *workspace.Index
	Settings    *lsp.ServerSettings
	LintOpts    analyzer.LintOptions
}

// getParsedFile returns the parsed stylesheet for a URI,
// protected by a read lock.
func (s *workspaceStore) getParsedFile(
	uri string,
) *parser.Stylesheet {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ParsedFiles[uri]
}

// getRawFile returns the raw source for a URI, protected by
// a read lock.
func (s *workspaceStore) getRawFile(
	uri string,
) []byte {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.RawFiles[uri]
}

const serverName = "go-css-lsp"

// modeFromString converts a setting string to a mode enum.
// Unrecognized values default to warn.
func modeFromString[T ~int](s string, ignore, err, warn T) T {
	switch s {
	case "ignore":
		return ignore
	case "error":
		return err
	default:
		return warn
	}
}

// TargetFileExtensions lists supported file extensions.
var TargetFileExtensions = []string{"css"}

// offsetRangeToLSPRange converts byte offsets to an LSP Range.
func offsetRangeToLSPRange(
	src []byte,
	start, end int,
) lsp.Range {
	startLine, startChar := css.OffsetToLineChar(src, start)
	endLine, endChar := css.OffsetToLineChar(src, end)
	return lsp.Range{
		Start: lsp.Position{
			Line:      uint(startLine), //nolint:gosec
			Character: uint(startChar), //nolint:gosec
		},
		End: lsp.Position{
			Line:      uint(endLine), //nolint:gosec
			Character: uint(endChar), //nolint:gosec
		},
	}
}

func main() {
	versionFlag := flag.Bool(
		"version", false, "print the LSP version",
	)
	flag.Parse()

	if *versionFlag {
		fmt.Printf(
			"%s -- version %s\n", serverName, version,
		)
		os.Exit(0)
	}

	configureLogging()
	scanner := lsp.ReceiveInput(os.Stdin)

	storage := &workspaceStore{
		RawFiles:    make(map[string][]byte),
		ParsedFiles: make(map[string]*parser.Stylesheet),
		VarIndex:    workspace.NewIndex(),
	}

	rootPathNotification := make(chan string, 2)
	textChangedNotification := make(chan bool, 2)
	textFromClient := make(map[string][]byte)
	muTextFromClient := new(sync.Mutex)

	go processDiagnosticNotification(
		storage,
		rootPathNotification,
		textChangedNotification,
		textFromClient,
		muTextFromClient,
	)

	var request lsp.RequestMessage[any]
	var response []byte
	var isRequestResponse bool
	var isExiting bool

	slog.Info("starting lsp server",
		slog.String("server_name", serverName),
		slog.String("server_version", version),
	)
	defer slog.Info("shutting down lsp server")

	for scanner.Scan() {
		data := scanner.Bytes()
		_ = json.Unmarshal(data, &request)

		if isExiting {
			if request.Method == lsp.MethodExit {
				break
			}
			response, _ = lsp.ProcessIllegalRequestAfterShutdown(
				request.JsonRpc, request.Id,
			)
			lsp.SendToLspClient(os.Stdout, response)
			continue
		}

		slog.Info("request " + request.Method)

		response, isRequestResponse = handleRequest(
			request.Method, data, &request,
			storage, rootPathNotification,
			textChangedNotification,
			textFromClient, muTextFromClient,
		)

		if request.Method == lsp.MethodInitialize {
			rootPathNotification = nil
		}
		if request.Method == lsp.MethodShutdown {
			isExiting = true
		}

		if isRequestResponse {
			lsp.SendToLspClient(os.Stdout, response)
		}
	}

	if scanner.Err() != nil {
		slog.Error(
			"error while closing LSP: " +
				scanner.Err().Error(),
		)
	}
}

// handleRequest dispatches a single LSP request with panic
// recovery. If any handler panics, the server logs the error
// and returns an internal error response instead of crashing.
func handleRequest(
	method string,
	data []byte,
	request *lsp.RequestMessage[any],
	storage *workspaceStore,
	rootPathNotification chan string,
	textChangedNotification chan bool,
	textFromClient map[string][]byte,
	mu *sync.Mutex,
) (response []byte, isRequestResponse bool) {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 4096)
			n := runtime.Stack(buf, false)
			slog.Error("panic in request handler",
				slog.String("method", method),
				slog.Any("panic", r),
				slog.String("stack", string(buf[:n])),
			)
			// Return an internal error response so the
			// client gets a reply instead of a dead server.
			response = lsp.MakeInternalError(
				request.JsonRpc, request.Id,
				fmt.Sprintf("internal error: %v", r),
			)
			isRequestResponse = true
		}
	}()

	switch method {
	case lsp.MethodInitialize:
		var rootURI string
		var settings *lsp.ServerSettings
		var err error
		response, rootURI, settings, err =
			lsp.ProcessInitializeRequest(
				data, serverName, version,
			)
		if err != nil {
			return nil, false
		}
		storage.Settings = settings
		if settings != nil {
			storage.LintOpts.Experimental = modeFromString(
				settings.ExperimentalFeatures,
				analyzer.ExperimentalIgnore,
				analyzer.ExperimentalError,
				analyzer.ExperimentalWarn,
			)
			storage.LintOpts.Deprecated = modeFromString(
				settings.DeprecatedFeatures,
				analyzer.DeprecatedIgnore,
				analyzer.DeprecatedError,
				analyzer.DeprecatedWarn,
			)
			storage.LintOpts.UnknownValues = modeFromString(
				settings.UnknownValues,
				analyzer.UnknownValueIgnore,
				analyzer.UnknownValueError,
				analyzer.UnknownValueWarn,
			)
			storage.LintOpts.StrictColorNames = settings.StrictColorNames
		}
		notifyTheRootPath(rootPathNotification, rootURI)
		return response, true

	case lsp.MethodInitialized:
		lsp.ProcessInitializedNotification(data)
		return nil, false

	case lsp.MethodShutdown:
		response, _ = lsp.ProcessShutdownRequest(
			request.JsonRpc, request.Id,
		)
		return response, true

	case lsp.MethodDidOpen:
		fileURI, fileContent, err :=
			lsp.ProcessDidOpenTextDocumentNotification(data)
		if err != nil {
			return nil, false
		}
		insertTextDocumentToDiagnostic(
			fileURI, fileContent,
			textChangedNotification,
			textFromClient, mu,
		)
		return nil, false

	case lsp.MethodDidChange:
		fileURI, fileContent, err :=
			lsp.ProcessDidChangeTextDocumentNotification(data)
		if err != nil {
			return nil, false
		}
		insertTextDocumentToDiagnostic(
			fileURI, fileContent,
			textChangedNotification,
			textFromClient, mu,
		)
		return nil, false

	case lsp.MethodDidClose:
		lsp.ProcessDidCloseTextDocumentNotification(data)
		return nil, false

	case lsp.MethodHover:
		return processHover(
			data, storage, textFromClient, mu,
		), true

	case lsp.MethodCompletion:
		return processCompletion(
			data, storage, textFromClient, mu,
		), true

	case lsp.MethodDocumentColor:
		return processDocumentColor(
			data, storage, textFromClient, mu,
		), true

	case lsp.MethodColorPresentation:
		return processColorPresentation(data), true

	case lsp.MethodSelectionRange:
		return processSelectionRange(
			data, storage, textFromClient, mu,
		), true

	case lsp.MethodPrepareRename:
		return processPrepareRename(
			data, storage, textFromClient, mu,
		), true

	case lsp.MethodRename:
		return processRename(
			data, storage, textFromClient, mu,
		), true

	case lsp.MethodFormatting:
		return processFormatting(
			data, storage, textFromClient, mu,
		), true

	case lsp.MethodDocumentHighlight:
		return processDocumentHighlight(
			data, storage, textFromClient, mu,
		), true

	case lsp.MethodFoldingRange:
		return processFoldingRange(
			data, storage, textFromClient, mu,
		), true

	case lsp.MethodDocumentLink:
		return processDocumentLink(
			data, storage, textFromClient, mu,
		), true

	case lsp.MethodCodeAction:
		return processCodeAction(
			data, storage, textFromClient, mu,
		), true

	case lsp.MethodReferences:
		return processReferences(
			data, storage, textFromClient, mu,
		), true

	case lsp.MethodDefinition:
		return processDefinition(
			data, storage, textFromClient, mu,
		), true

	case lsp.MethodDocumentSymbol:
		return processDocumentSymbol(
			data, storage, textFromClient, mu,
		), true

	default:
		return nil, false
	}
}

// processHover handles textDocument/hover requests.
func processHover(
	data []byte,
	storage *workspaceStore,
	textFromClient map[string][]byte,
	mu *sync.Mutex,
) []byte {
	var req lsp.RequestMessage[lsp.HoverParams]
	if err := json.Unmarshal(data, &req); err != nil {
		slog.Warn(
			"Error unmarshalling hover request: " +
				err.Error(),
		)
		return nil
	}

	uri := req.Params.TextDocument.Uri
	src := getFileContent(uri, storage, textFromClient, mu)
	if src == nil {
		return marshalNullResult(req.JsonRpc, req.Id)
	}

	ss := storage.getParsedFile(uri)
	if ss == nil {
		result := css.Parse(src)
		ss = result.Stylesheet
	}

	hover := css.Hover(
		ss, src,
		int(req.Params.Position.Line),      //nolint:gosec // LSP positions are small
		int(req.Params.Position.Character), //nolint:gosec // LSP positions are small
		storage.VarIndex,
	)

	res := lsp.ResponseMessage[*lsp.HoverResult]{
		JsonRpc: req.JsonRpc,
		Id:      req.Id,
	}

	if hover.Found {
		hr := &lsp.HoverResult{
			Contents: lsp.MarkupContent{
				Kind:  "markdown",
				Value: hover.Content,
			},
		}
		if hover.RangeStart < hover.RangeEnd {
			r := offsetRangeToLSPRange(
				src, hover.RangeStart, hover.RangeEnd,
			)
			hr.Range = &r
		}
		res.Result = hr
	}

	out, err := json.Marshal(res)
	if err != nil {
		slog.Warn(
			"Error marshalling hover response: " +
				err.Error(),
		)
		return nil
	}
	return out
}

// processCompletion handles textDocument/completion requests.
func processCompletion(
	data []byte,
	storage *workspaceStore,
	textFromClient map[string][]byte,
	mu *sync.Mutex,
) []byte {
	var req lsp.RequestMessage[lsp.CompletionParams]
	if err := json.Unmarshal(data, &req); err != nil {
		slog.Warn(
			"Error unmarshalling completion request: " +
				err.Error(),
		)
		return nil
	}

	uri := req.Params.TextDocument.Uri
	src := getFileContent(uri, storage, textFromClient, mu)
	if src == nil {
		return marshalNullResult(req.JsonRpc, req.Id)
	}

	ss := storage.getParsedFile(uri)
	if ss == nil {
		result := css.Parse(src)
		ss = result.Stylesheet
	}

	items := css.Completions(
		ss, src,
		int(req.Params.Position.Line),      //nolint:gosec // LSP positions are small
		int(req.Params.Position.Character), //nolint:gosec // LSP positions are small
		storage.LintOpts,
	)

	lspItems := make([]lsp.CompletionItem, len(items))
	for i, item := range items {
		lspItems[i] = lsp.CompletionItem{
			Label:      item.Label,
			Kind:       item.Kind,
			Detail:     item.Detail,
			InsertText: item.InsertText,
			Tags:       item.Tags,
			Deprecated: item.Deprecated,
		}
		if item.Documentation != "" {
			lspItems[i].Documentation = &lsp.MarkupContent{
				Kind:  "markdown",
				Value: item.Documentation,
			}
		}
	}

	res := lsp.ResponseMessage[*lsp.CompletionList]{
		JsonRpc: req.JsonRpc,
		Id:      req.Id,
		Result: &lsp.CompletionList{
			IsIncomplete: false,
			Items:        lspItems,
		},
	}

	out, err := json.Marshal(res)
	if err != nil {
		slog.Warn(
			"Error marshalling completion response: " +
				err.Error(),
		)
		return nil
	}
	return out
}

// processDocumentColor handles textDocument/documentColor.
func processDocumentColor(
	data []byte,
	storage *workspaceStore,
	textFromClient map[string][]byte,
	mu *sync.Mutex,
) []byte {
	var req lsp.RequestMessage[lsp.DocumentColorParams]
	if err := json.Unmarshal(data, &req); err != nil {
		slog.Warn(
			"Error unmarshalling documentColor request: " +
				err.Error(),
		)
		return nil
	}

	uri := req.Params.TextDocument.Uri
	src := getFileContent(uri, storage, textFromClient, mu)
	if src == nil {
		return marshalEmptyArray(req.JsonRpc, req.Id)
	}

	ss := storage.getParsedFile(uri)
	if ss == nil {
		result := css.Parse(src)
		ss = result.Stylesheet
	}

	docColors := css.DocumentColorsResolved(
		ss, src, storage.VarIndex,
	)
	result := make([]lsp.ColorInformation, len(docColors))

	for i, dc := range docColors {
		result[i] = lsp.ColorInformation{
			Range: offsetRangeToLSPRange(
				src, dc.StartPos, dc.EndPos,
			),
			Color: lsp.LSPColor{
				Red:   dc.Color.Red,
				Green: dc.Color.Green,
				Blue:  dc.Color.Blue,
				Alpha: dc.Color.Alpha,
			},
		}
	}

	res := lsp.ResponseMessage[[]lsp.ColorInformation]{
		JsonRpc: req.JsonRpc,
		Id:      req.Id,
		Result:  result,
	}

	out, err := json.Marshal(res)
	if err != nil {
		slog.Warn(
			"Error marshalling documentColor response: " +
				err.Error(),
		)
		return nil
	}
	return out
}

// processColorPresentation handles
// textDocument/colorPresentation.
func processColorPresentation(data []byte) []byte {
	var req lsp.RequestMessage[lsp.ColorPresentationParams]
	if err := json.Unmarshal(data, &req); err != nil {
		slog.Warn(
			"Error unmarshalling colorPresentation request: " +
				err.Error(),
		)
		return nil
	}

	c := analyzer.Color{
		Red:   req.Params.Color.Red,
		Green: req.Params.Color.Green,
		Blue:  req.Params.Color.Blue,
		Alpha: req.Params.Color.Alpha,
	}

	labels := css.ColorPresentations(c)
	result := make([]lsp.ColorPresentation, len(labels))
	for i, label := range labels {
		result[i] = lsp.ColorPresentation{
			Label: label,
			TextEdit: &lsp.TextEdit{
				Range:   req.Params.Range,
				NewText: label,
			},
		}
	}

	res := lsp.ResponseMessage[[]lsp.ColorPresentation]{
		JsonRpc: req.JsonRpc,
		Id:      req.Id,
		Result:  result,
	}

	out, err := json.Marshal(res)
	if err != nil {
		slog.Warn(
			"Error marshalling colorPresentation response: " +
				err.Error(),
		)
		return nil
	}
	return out
}

// processSelectionRange handles textDocument/selectionRange.
func processSelectionRange(
	data []byte,
	storage *workspaceStore,
	textFromClient map[string][]byte,
	mu *sync.Mutex,
) []byte {
	var req lsp.RequestMessage[lsp.SelectionRangeParams]
	if err := json.Unmarshal(data, &req); err != nil {
		slog.Warn(
			"Error unmarshalling selectionRange request: " +
				err.Error(),
		)
		return nil
	}

	uri := req.Params.TextDocument.Uri
	src := getFileContent(uri, storage, textFromClient, mu)
	if src == nil {
		return marshalEmptyArray(req.JsonRpc, req.Id)
	}

	ss := storage.getParsedFile(uri)
	if ss == nil {
		result := css.Parse(src)
		ss = result.Stylesheet
	}

	result := make([]lsp.LSPSelectionRange, len(req.Params.Positions))
	for i, pos := range req.Params.Positions {
		sr := css.SelectionRange(
			ss, src,
			int(pos.Line),      //nolint:gosec
			int(pos.Character), //nolint:gosec
		)
		result[i] = convertSelectionRange(sr, src)
	}

	res := lsp.ResponseMessage[[]lsp.LSPSelectionRange]{
		JsonRpc: req.JsonRpc,
		Id:      req.Id,
		Result:  result,
	}

	out, err := json.Marshal(res)
	if err != nil {
		slog.Warn(
			"Error marshalling selectionRange response: " +
				err.Error(),
		)
		return nil
	}
	return out
}

func convertSelectionRange(
	sr *analyzer.SelectionRange,
	src []byte,
) lsp.LSPSelectionRange {
	if sr == nil {
		return lsp.LSPSelectionRange{}
	}

	result := lsp.LSPSelectionRange{
		Range: offsetRangeToLSPRange(
			src, sr.StartPos, sr.EndPos,
		),
	}

	if sr.Parent != nil {
		parent := convertSelectionRange(sr.Parent, src)
		result.Parent = &parent
	}

	return result
}

// processPrepareRename handles textDocument/prepareRename.
func processPrepareRename(
	data []byte,
	storage *workspaceStore,
	textFromClient map[string][]byte,
	mu *sync.Mutex,
) []byte {
	var req lsp.RequestMessage[lsp.TextDocumentPositionParams]
	if err := json.Unmarshal(data, &req); err != nil {
		slog.Warn(
			"Error unmarshalling prepareRename request: " +
				err.Error(),
		)
		return nil
	}

	uri := req.Params.TextDocument.Uri
	src := getFileContent(uri, storage, textFromClient, mu)
	if src == nil {
		return marshalNullResult(req.JsonRpc, req.Id)
	}

	ss := storage.getParsedFile(uri)
	if ss == nil {
		result := css.Parse(src)
		ss = result.Stylesheet
	}

	loc, found := css.PrepareRename(
		ss, src,
		int(req.Params.Position.Line),      //nolint:gosec
		int(req.Params.Position.Character), //nolint:gosec
	)

	if !found {
		return marshalNullResult(req.JsonRpc, req.Id)
	}

	result := offsetRangeToLSPRange(
		src, loc.StartPos, loc.EndPos,
	)

	res := lsp.ResponseMessage[lsp.Range]{
		JsonRpc: req.JsonRpc,
		Id:      req.Id,
		Result:  result,
	}

	out, err := json.Marshal(res)
	if err != nil {
		slog.Warn(
			"Error marshalling prepareRename response: " +
				err.Error(),
		)
		return nil
	}
	return out
}

// processRename handles textDocument/rename.
func processRename(
	data []byte,
	storage *workspaceStore,
	textFromClient map[string][]byte,
	mu *sync.Mutex,
) []byte {
	var req lsp.RequestMessage[lsp.RenameParams]
	if err := json.Unmarshal(data, &req); err != nil {
		slog.Warn(
			"Error unmarshalling rename request: " +
				err.Error(),
		)
		return nil
	}

	uri := req.Params.TextDocument.Uri
	src := getFileContent(uri, storage, textFromClient, mu)
	if src == nil {
		return marshalNullResult(req.JsonRpc, req.Id)
	}

	ss := storage.getParsedFile(uri)
	if ss == nil {
		result := css.Parse(src)
		ss = result.Stylesheet
	}

	edits := css.Rename(
		ss, src,
		int(req.Params.Position.Line),      //nolint:gosec
		int(req.Params.Position.Character), //nolint:gosec
		req.Params.NewName,
	)

	if len(edits) == 0 {
		return marshalNullResult(req.JsonRpc, req.Id)
	}

	textEdits := make([]lsp.TextEdit, len(edits))
	for i, e := range edits {
		textEdits[i] = lsp.TextEdit{
			Range: offsetRangeToLSPRange(
				src, e.StartPos, e.EndPos,
			),
			NewText: e.NewText,
		}
	}

	result := lsp.WorkspaceEdit{
		Changes: map[string][]lsp.TextEdit{
			uri: textEdits,
		},
	}

	res := lsp.ResponseMessage[lsp.WorkspaceEdit]{
		JsonRpc: req.JsonRpc,
		Id:      req.Id,
		Result:  result,
	}

	out, err := json.Marshal(res)
	if err != nil {
		slog.Warn(
			"Error marshalling rename response: " +
				err.Error(),
		)
		return nil
	}
	return out
}

// processFormatting handles textDocument/formatting.
func processFormatting(
	data []byte,
	storage *workspaceStore,
	textFromClient map[string][]byte,
	mu *sync.Mutex,
) []byte {
	var req lsp.RequestMessage[lsp.DocumentFormattingParams]
	if err := json.Unmarshal(data, &req); err != nil {
		slog.Warn(
			"Error unmarshalling formatting request: " +
				err.Error(),
		)
		return nil
	}

	uri := req.Params.TextDocument.Uri
	src := getFileContent(uri, storage, textFromClient, mu)
	if src == nil {
		return marshalEmptyArray(req.JsonRpc, req.Id)
	}

	ss := storage.getParsedFile(uri)
	if ss == nil {
		result := css.Parse(src)
		ss = result.Stylesheet
	}

	fmtOpts := analyzer.FormatOptions{
		TabSize:      req.Params.Options.TabSize,
		InsertSpaces: req.Params.Options.InsertSpaces,
	}
	if storage.Settings != nil {
		switch storage.Settings.FormatMode {
		case "compact":
			fmtOpts.Mode = analyzer.FormatCompact
		case "preserve":
			fmtOpts.Mode = analyzer.FormatPreserve
		case "detect":
			fmtOpts.Mode = analyzer.FormatDetect
		}
		fmtOpts.PrintWidth = storage.Settings.PrintWidth
	}

	formatted := css.FormatDocument(ss, src, fmtOpts)

	result := []lsp.TextEdit{
		{
			Range: offsetRangeToLSPRange(
				src, 0, len(src),
			),
			NewText: formatted,
		},
	}

	res := lsp.ResponseMessage[[]lsp.TextEdit]{
		JsonRpc: req.JsonRpc,
		Id:      req.Id,
		Result:  result,
	}

	out, err := json.Marshal(res)
	if err != nil {
		slog.Warn(
			"Error marshalling formatting response: " +
				err.Error(),
		)
		return nil
	}
	return out
}

// processDocumentHighlight handles textDocument/documentHighlight.
func processDocumentHighlight(
	data []byte,
	storage *workspaceStore,
	textFromClient map[string][]byte,
	mu *sync.Mutex,
) []byte {
	var req lsp.RequestMessage[lsp.TextDocumentPositionParams]
	if err := json.Unmarshal(data, &req); err != nil {
		slog.Warn(
			"Error unmarshalling documentHighlight request: " +
				err.Error(),
		)
		return nil
	}

	uri := req.Params.TextDocument.Uri
	src := getFileContent(uri, storage, textFromClient, mu)
	if src == nil {
		return marshalEmptyArray(req.JsonRpc, req.Id)
	}

	ss := storage.getParsedFile(uri)
	if ss == nil {
		result := css.Parse(src)
		ss = result.Stylesheet
	}

	highlights := css.DocumentHighlights(
		ss, src,
		int(req.Params.Position.Line),      //nolint:gosec
		int(req.Params.Position.Character), //nolint:gosec
	)

	result := make([]lsp.LSPDocumentHighlight, len(highlights))
	for i, h := range highlights {
		result[i] = lsp.LSPDocumentHighlight{
			Range: offsetRangeToLSPRange(
				src, h.StartPos, h.EndPos,
			),
			Kind: h.Kind,
		}
	}

	res := lsp.ResponseMessage[[]lsp.LSPDocumentHighlight]{
		JsonRpc: req.JsonRpc,
		Id:      req.Id,
		Result:  result,
	}

	out, err := json.Marshal(res)
	if err != nil {
		slog.Warn(
			"Error marshalling documentHighlight response: " +
				err.Error(),
		)
		return nil
	}
	return out
}

// processFoldingRange handles textDocument/foldingRange.
func processFoldingRange(
	data []byte,
	storage *workspaceStore,
	textFromClient map[string][]byte,
	mu *sync.Mutex,
) []byte {
	var req lsp.RequestMessage[lsp.FoldingRangeParams]
	if err := json.Unmarshal(data, &req); err != nil {
		slog.Warn(
			"Error unmarshalling foldingRange request: " +
				err.Error(),
		)
		return nil
	}

	uri := req.Params.TextDocument.Uri
	src := getFileContent(uri, storage, textFromClient, mu)
	if src == nil {
		return marshalEmptyArray(req.JsonRpc, req.Id)
	}

	ss := storage.getParsedFile(uri)
	if ss == nil {
		result := css.Parse(src)
		ss = result.Stylesheet
	}

	ranges := css.FoldingRanges(ss, src)
	result := make([]lsp.LSPFoldingRange, len(ranges))
	for i, r := range ranges {
		startLine, _ := css.OffsetToLineChar(src, r.StartPos)
		endLine, _ := css.OffsetToLineChar(src, r.EndPos)
		result[i] = lsp.LSPFoldingRange{
			StartLine: startLine,
			EndLine:   endLine,
			Kind:      r.Kind,
		}
	}

	res := lsp.ResponseMessage[[]lsp.LSPFoldingRange]{
		JsonRpc: req.JsonRpc,
		Id:      req.Id,
		Result:  result,
	}

	out, err := json.Marshal(res)
	if err != nil {
		slog.Warn(
			"Error marshalling foldingRange response: " +
				err.Error(),
		)
		return nil
	}
	return out
}

// processDocumentLink handles textDocument/documentLink.
func processDocumentLink(
	data []byte,
	storage *workspaceStore,
	textFromClient map[string][]byte,
	mu *sync.Mutex,
) []byte {
	var req lsp.RequestMessage[lsp.DocumentLinkParams]
	if err := json.Unmarshal(data, &req); err != nil {
		slog.Warn(
			"Error unmarshalling documentLink request: " +
				err.Error(),
		)
		return nil
	}

	uri := req.Params.TextDocument.Uri
	src := getFileContent(uri, storage, textFromClient, mu)
	if src == nil {
		return marshalEmptyArray(req.JsonRpc, req.Id)
	}

	ss := storage.getParsedFile(uri)
	if ss == nil {
		result := css.Parse(src)
		ss = result.Stylesheet
	}

	links := css.DocumentLinks(ss, src)
	result := make([]lsp.LSPDocumentLink, len(links))
	for i, l := range links {
		result[i] = lsp.LSPDocumentLink{
			Range: offsetRangeToLSPRange(
				src, l.StartPos, l.EndPos,
			),
			Target: l.Target,
		}
	}

	res := lsp.ResponseMessage[[]lsp.LSPDocumentLink]{
		JsonRpc: req.JsonRpc,
		Id:      req.Id,
		Result:  result,
	}

	out, err := json.Marshal(res)
	if err != nil {
		slog.Warn(
			"Error marshalling documentLink response: " +
				err.Error(),
		)
		return nil
	}
	return out
}

// containsOnly checks if a code action kind is in the only
// filter list.
func containsOnly(only []string, kind string) bool {
	return slices.Contains(only, kind)
}

// processFixAll computes all fixable edits and returns them as
// a single source.fixAll code action.
func processFixAll(
	req lsp.RequestMessage[lsp.CodeActionParams],
	src []byte,
	uri string,
	storage *workspaceStore,
) []byte {
	actions := css.FixAllActions(src, storage.LintOpts)
	if len(actions) == 0 {
		return marshalEmptyArray(req.JsonRpc, req.Id)
	}

	edits := make([]lsp.TextEdit, len(actions))
	for i, a := range actions {
		edits[i] = lsp.TextEdit{
			Range: lsp.Range{
				Start: lsp.Position{
					Line:      uint(a.StartLine), //nolint:gosec
					Character: uint(a.StartChar), //nolint:gosec
				},
				End: lsp.Position{
					Line:      uint(a.EndLine), //nolint:gosec
					Character: uint(a.EndChar), //nolint:gosec
				},
			},
			NewText: a.ReplaceWith,
		}
	}

	result := []lsp.LSPCodeAction{{
		Title: "Fix all auto-fixable problems",
		Kind:  analyzer.CodeActionSourceFixAll,
		Edit: &lsp.WorkspaceEdit{
			Changes: map[string][]lsp.TextEdit{
				uri: edits,
			},
		},
	}}

	res := lsp.ResponseMessage[[]lsp.LSPCodeAction]{
		JsonRpc: req.JsonRpc,
		Id:      req.Id,
		Result:  result,
	}

	out, err := json.Marshal(res)
	if err != nil {
		slog.Warn(
			"Error marshalling fixAll response: " +
				err.Error(),
		)
		return nil
	}
	return out
}

// processCodeAction handles textDocument/codeAction.
func processCodeAction(
	data []byte,
	storage *workspaceStore,
	textFromClient map[string][]byte,
	mu *sync.Mutex,
) []byte {
	var req lsp.RequestMessage[lsp.CodeActionParams]
	if err := json.Unmarshal(data, &req); err != nil {
		slog.Warn(
			"Error unmarshalling codeAction request: " +
				err.Error(),
		)
		return nil
	}

	uri := req.Params.TextDocument.Uri
	src := getFileContent(uri, storage, textFromClient, mu)
	if src == nil {
		return marshalEmptyArray(req.JsonRpc, req.Id)
	}

	// Handle source.fixAll requests
	if containsOnly(
		req.Params.Context.Only,
		analyzer.CodeActionSourceFixAll,
	) {
		return processFixAll(req, src, uri, storage)
	}

	// Convert LSP diagnostics to analyzer diagnostics
	var analyzerDiags []analyzer.Diagnostic
	for _, d := range req.Params.Context.Diagnostics {
		analyzerDiags = append(analyzerDiags,
			analyzer.Diagnostic{
				Message:   d.Message,
				StartLine: int(d.Range.Start.Line),      //nolint:gosec
				StartChar: int(d.Range.Start.Character), //nolint:gosec
				EndLine:   int(d.Range.End.Line),        //nolint:gosec
				EndChar:   int(d.Range.End.Character),   //nolint:gosec
				Severity:  d.Severity,
			},
		)
	}

	mu.Lock()
	ss := storage.getParsedFile(uri)
	mu.Unlock()
	if ss == nil {
		result := css.Parse(src)
		ss = result.Stylesheet
	}

	cursorLine := int(req.Params.Range.Start.Line)      //nolint:gosec
	cursorChar := int(req.Params.Range.Start.Character) //nolint:gosec

	actions := css.CodeActions(
		ss, src, cursorLine, cursorChar, analyzerDiags,
	)
	result := make([]lsp.LSPCodeAction, len(actions))
	for i, a := range actions {
		result[i] = lsp.LSPCodeAction{
			Title: a.Title,
			Kind:  a.Kind,
			Edit: &lsp.WorkspaceEdit{
				Changes: map[string][]lsp.TextEdit{
					uri: {
						{
							Range: lsp.Range{
								Start: lsp.Position{
									Line:      uint(a.StartLine), //nolint:gosec
									Character: uint(a.StartChar), //nolint:gosec
								},
								End: lsp.Position{
									Line:      uint(a.EndLine), //nolint:gosec
									Character: uint(a.EndChar), //nolint:gosec
								},
							},
							NewText: a.ReplaceWith,
						},
					},
				},
			},
		}
	}

	res := lsp.ResponseMessage[[]lsp.LSPCodeAction]{
		JsonRpc: req.JsonRpc,
		Id:      req.Id,
		Result:  result,
	}

	out, err := json.Marshal(res)
	if err != nil {
		slog.Warn(
			"Error marshalling codeAction response: " +
				err.Error(),
		)
		return nil
	}
	return out
}

// processReferences handles textDocument/references.
func processReferences(
	data []byte,
	storage *workspaceStore,
	textFromClient map[string][]byte,
	mu *sync.Mutex,
) []byte {
	var req lsp.RequestMessage[lsp.ReferenceParams]
	if err := json.Unmarshal(data, &req); err != nil {
		slog.Warn(
			"Error unmarshalling references request: " +
				err.Error(),
		)
		return nil
	}

	uri := req.Params.TextDocument.Uri
	src := getFileContent(uri, storage, textFromClient, mu)
	if src == nil {
		return marshalEmptyArray(req.JsonRpc, req.Id)
	}

	ss := storage.getParsedFile(uri)
	if ss == nil {
		result := css.Parse(src)
		ss = result.Stylesheet
	}

	refs := css.References(
		ss, src,
		int(req.Params.Position.Line),      //nolint:gosec
		int(req.Params.Position.Character), //nolint:gosec
	)

	result := make([]lsp.LSPLocation, len(refs))
	for i, ref := range refs {
		result[i] = lsp.LSPLocation{
			URI: uri,
			Range: offsetRangeToLSPRange(
				src, ref.StartPos, ref.EndPos,
			),
		}
	}

	res := lsp.ResponseMessage[[]lsp.LSPLocation]{
		JsonRpc: req.JsonRpc,
		Id:      req.Id,
		Result:  result,
	}

	out, err := json.Marshal(res)
	if err != nil {
		slog.Warn(
			"Error marshalling references response: " +
				err.Error(),
		)
		return nil
	}
	return out
}

// processDefinition handles textDocument/definition.
func processDefinition(
	data []byte,
	storage *workspaceStore,
	textFromClient map[string][]byte,
	mu *sync.Mutex,
) []byte {
	var req lsp.RequestMessage[lsp.TextDocumentPositionParams]
	if err := json.Unmarshal(data, &req); err != nil {
		slog.Warn(
			"Error unmarshalling definition request: " +
				err.Error(),
		)
		return nil
	}

	uri := req.Params.TextDocument.Uri
	src := getFileContent(uri, storage, textFromClient, mu)
	if src == nil {
		return marshalNullResult(req.JsonRpc, req.Id)
	}

	ss := storage.getParsedFile(uri)
	if ss == nil {
		result := css.Parse(src)
		ss = result.Stylesheet
	}

	defResult, found := css.Definition(
		ss, src,
		int(req.Params.Position.Line),      //nolint:gosec
		int(req.Params.Position.Character), //nolint:gosec
	)

	if found {
		originRange := offsetRangeToLSPRange(
			src, defResult.OriginStart, defResult.OriginEnd,
		)
		targetRange := offsetRangeToLSPRange(
			src, defResult.TargetStart, defResult.TargetEnd,
		)
		result := []lsp.LSPLocationLink{{
			OriginSelectionRange: &originRange,
			TargetUri:            uri,
			TargetRange:          targetRange,
			TargetSelectionRange: targetRange,
		}}
		return marshalDefinitionResult(
			req.JsonRpc, req.Id, result,
		)
	}

	// Fall back to workspace index for cross-file lookup
	varName, originStart, originEnd := css.VarReferenceWithRange(
		ss, src,
		int(req.Params.Position.Line),      //nolint:gosec
		int(req.Params.Position.Character), //nolint:gosec
	)
	if varName == "" {
		return marshalNullResult(req.JsonRpc, req.Id)
	}

	defs := storage.VarIndex.LookupDefinitions(varName)
	if len(defs) == 0 {
		return marshalNullResult(req.JsonRpc, req.Id)
	}

	def := defs[0]
	targetSrc := storage.getRawFile(def.URI)
	if targetSrc == nil {
		return marshalNullResult(req.JsonRpc, req.Id)
	}

	originRange := offsetRangeToLSPRange(
		src, originStart, originEnd,
	)
	targetRange := offsetRangeToLSPRange(
		targetSrc, def.StartPos, def.EndPos,
	)
	result := []lsp.LSPLocationLink{{
		OriginSelectionRange: &originRange,
		TargetUri:            def.URI,
		TargetRange:          targetRange,
		TargetSelectionRange: targetRange,
	}}
	return marshalDefinitionResult(
		req.JsonRpc, req.Id, result,
	)
}

func marshalDefinitionResult(
	jsonRpc string,
	id lsp.ID,
	result []lsp.LSPLocationLink,
) []byte {
	res := lsp.ResponseMessage[[]lsp.LSPLocationLink]{
		JsonRpc: jsonRpc,
		Id:      id,
		Result:  result,
	}
	out, err := json.Marshal(res)
	if err != nil {
		slog.Warn(
			"Error marshalling definition response: " +
				err.Error(),
		)
		return nil
	}
	return out
}

// processDocumentSymbol handles textDocument/documentSymbol.
func processDocumentSymbol(
	data []byte,
	storage *workspaceStore,
	textFromClient map[string][]byte,
	mu *sync.Mutex,
) []byte {
	var req lsp.RequestMessage[lsp.DocumentSymbolParams]
	if err := json.Unmarshal(data, &req); err != nil {
		slog.Warn(
			"Error unmarshalling documentSymbol request: " +
				err.Error(),
		)
		return nil
	}

	uri := req.Params.TextDocument.Uri
	src := getFileContent(uri, storage, textFromClient, mu)
	if src == nil {
		return marshalEmptyArray(req.JsonRpc, req.Id)
	}

	ss := storage.getParsedFile(uri)
	if ss == nil {
		result := css.Parse(src)
		ss = result.Stylesheet
	}

	symbols := css.DocumentSymbols(ss, src)
	result := convertSymbols(symbols, src)

	res := lsp.ResponseMessage[[]lsp.LSPDocumentSymbol]{
		JsonRpc: req.JsonRpc,
		Id:      req.Id,
		Result:  result,
	}

	out, err := json.Marshal(res)
	if err != nil {
		slog.Warn(
			"Error marshalling documentSymbol response: " +
				err.Error(),
		)
		return nil
	}
	return out
}

func convertSymbols(
	symbols []analyzer.DocumentSymbol,
	src []byte,
) []lsp.LSPDocumentSymbol {
	result := make([]lsp.LSPDocumentSymbol, len(symbols))
	for i, s := range symbols {
		result[i] = lsp.LSPDocumentSymbol{
			Name: s.Name,
			Kind: s.Kind,
			Range: offsetRangeToLSPRange(
				src, s.StartPos, s.EndPos,
			),
			SelectionRange: offsetRangeToLSPRange(
				src, s.SelectionStart, s.SelectionEnd,
			),
		}

		if len(s.Children) > 0 {
			result[i].Children = convertSymbols(
				s.Children, src,
			)
		}
	}
	return result
}

func marshalEmptyArray(
	jsonRpc string,
	id lsp.ID,
) []byte {
	res := lsp.ResponseMessage[[]any]{
		JsonRpc: jsonRpc,
		Id:      id,
		Result:  []any{},
	}
	out, _ := json.Marshal(res)
	return out
}

func getFileContent(
	uri string,
	storage *workspaceStore,
	textFromClient map[string][]byte,
	mu *sync.Mutex,
) []byte {
	mu.Lock()
	content := textFromClient[uri]
	mu.Unlock()

	if content != nil {
		return content
	}

	storage.mu.RLock()
	defer storage.mu.RUnlock()
	return storage.RawFiles[uri]
}

func marshalNullResult(
	jsonRpc string,
	id lsp.ID,
) []byte {
	res := lsp.ResponseMessage[any]{
		JsonRpc: jsonRpc,
		Id:      id,
	}
	out, _ := json.Marshal(res)
	return out
}

// insertTextDocumentToDiagnostic queues a document for
// processing.
func insertTextDocumentToDiagnostic(
	uri string,
	content []byte,
	textChangedNotification chan bool,
	textFromClient map[string][]byte,
	muTextFromClient *sync.Mutex,
) {
	if uri == "" {
		return
	}

	muTextFromClient.Lock()
	textFromClient[uri] = content

	if len(textChangedNotification) == 0 {
		textChangedNotification <- true
	}

	muTextFromClient.Unlock()
}

// notifyTheRootPath sends the root path to the diagnostic
// goroutine.
func notifyTheRootPath(
	rootPathNotification chan string,
	rootURI string,
) {
	if rootPathNotification == nil {
		return
	}

	rootPathNotification <- rootURI
	close(rootPathNotification)
}

// processDiagnosticNotification runs diagnostics and sends
// notifications to the client.
func processDiagnosticNotification(
	storage *workspaceStore,
	rootPathNotification chan string,
	textChangedNotification chan bool,
	textFromClient map[string][]byte,
	muTextFromClient *sync.Mutex,
) {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 4096)
			n := runtime.Stack(buf, false)
			slog.Error("panic in diagnostic goroutine",
				slog.Any("panic", r),
				slog.String("stack", string(buf[:n])),
			)
		}
	}()

	rootURI, ok := <-rootPathNotification
	if !ok {
		return
	}

	storage.RootPath = uriToFilePath(rootURI)

	// Scan workspace for CSS variables
	if storage.RootPath != "" {
		if err := storage.VarIndex.ScanWorkspace(
			storage.RootPath,
		); err != nil {
			slog.Warn(
				"workspace scan error: " + err.Error(),
			)
		}
	}

	notification := &lsp.NotificationMessage[lsp.PublishDiagnosticsParams]{
		JsonRpc: lsp.JSONRPCVersion,
		Method:  lsp.MethodPublishDiagnostics,
		Params: lsp.PublishDiagnosticsParams{
			Uri:         "",
			Diagnostics: []lsp.Diagnostic{},
		},
	}

	for range textChangedNotification {
		muTextFromClient.Lock()

		filesToProcess := make(map[string][]byte)
		maps.Copy(filesToProcess, textFromClient)
		clear(textFromClient)

		// Drain extra notifications
		for range len(textChangedNotification) {
			<-textChangedNotification
		}

		muTextFromClient.Unlock()

		for uri, content := range filesToProcess {
			diags, ss := css.Diagnostics(content, storage.LintOpts)

			storage.mu.Lock()
			storage.RawFiles[uri] = content
			storage.ParsedFiles[uri] = ss
			storage.mu.Unlock()

			// Update workspace variable index using
			// pre-parsed stylesheet to avoid double-parse
			storage.VarIndex.IndexFileWithStylesheet(uri, ss, content)

			notification.Params.Uri = uri
			notification.Params.Diagnostics = convertDiagnostics(diags)

			out, err := json.Marshal(notification)
			if err != nil {
				slog.Error(
					"unable to marshal notification: " +
						err.Error(),
				)
				continue
			}

			lsp.SendToLspClient(os.Stdout, out)
		}
	}
}

func convertDiagnostics(
	diags []analyzer.Diagnostic,
) []lsp.Diagnostic {
	result := make([]lsp.Diagnostic, len(diags))
	for i, d := range diags {
		result[i] = lsp.Diagnostic{
			Range: lsp.Range{
				Start: lsp.Position{
					Line:      uint(d.StartLine), //nolint:gosec
					Character: uint(d.StartChar), //nolint:gosec
				},
				End: lsp.Position{
					Line:      uint(d.EndLine), //nolint:gosec
					Character: uint(d.EndChar), //nolint:gosec
				},
			},
			Message:  d.Message,
			Severity: d.Severity,
		}
	}
	return result
}

// uriToFilePath converts a file URI to an OS path.
func uriToFilePath(uri string) string {
	if uri == "" {
		return ""
	}

	u, err := url.Parse(uri)
	if err != nil {
		slog.Error(
			"unable to parse URI: " + err.Error(),
		)
		return ""
	}

	path := u.Path
	if runtime.GOOS == "windows" {
		if path[0] == '/' && len(path) >= 3 &&
			path[2] == ':' {
			path = path[1:]
		}
	}

	return filepath.FromSlash(path)
}

// createLogFile creates or opens the log file.
func createLogFile() *os.File {
	userCachePath, err := os.UserCacheDir()
	if err != nil {
		return os.Stderr
	}

	appCachePath := filepath.Join(
		userCachePath, "go-css-lsp",
	)
	logFilePath := filepath.Join(
		appCachePath, "go-css-lsp.log",
	)

	_ = os.Mkdir(appCachePath, lsp.DirPermissions)

	fileInfo, err := os.Stat(logFilePath)
	if err == nil && fileInfo.Size() >= lsp.MaxLogFileSize {
		//nolint:gosec // safe log file path
		file, err := os.OpenFile(
			logFilePath,
			os.O_TRUNC|os.O_WRONLY,
			lsp.FilePermissions,
		)
		if err != nil {
			return os.Stderr
		}
		return file
	}

	//nolint:gosec // safe log file path
	file, err := os.OpenFile(
		logFilePath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		lsp.FilePermissions,
	)
	if err != nil {
		return os.Stderr
	}

	return file
}

// configureLogging sets up structured logging.
func configureLogging() {
	file := createLogFile()
	if file == nil {
		file = os.Stderr
	}

	logger := slog.New(slog.NewJSONHandler(file, nil))
	slog.SetDefault(logger)
}
