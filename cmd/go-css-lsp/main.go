// Command go-css-lsp provides a Language Server Protocol server
// for CSS.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"slices"
	"sync"

	"github.com/toba/css-lsp/internal/css"
	"github.com/toba/css-lsp/internal/css/analyzer"
	"github.com/toba/css-lsp/internal/css/parser"
	"github.com/toba/css-lsp/internal/css/workspace"
	"github.com/toba/lsp/pathutil"
	"github.com/toba/lsp/server"
	"go.lsp.dev/protocol"
)

// version is set by goreleaser at build time.
var version = "dev"

const serverName = "go-css-lsp"

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

// cssHandler implements server.Handler and optional handler
// interfaces for the CSS language server.
type cssHandler struct {
	mu          sync.RWMutex
	rootPath    string
	rawFiles    map[string][]byte
	parsedFiles map[string]*parser.Stylesheet
	varIndex    *workspace.Index
	settings    *ServerSettings
	lintOpts    analyzer.LintOptions
}

func newCSSHandler() *cssHandler {
	return &cssHandler{
		rawFiles:    make(map[string][]byte),
		parsedFiles: make(map[string]*parser.Stylesheet),
		varIndex:    workspace.NewIndex(),
	}
}

// modeFromString converts a setting string to a mode enum.
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

// offsetRangeToProtocolRange converts byte offsets to a
// protocol.Range.
func offsetRangeToProtocolRange(
	src []byte,
	start, end int,
) protocol.Range {
	startLine, startChar := css.OffsetToLineChar(src, start)
	endLine, endChar := css.OffsetToLineChar(src, end)
	return protocol.Range{
		Start: protocol.Position{
			Line:      uint32(startLine), //nolint:gosec
			Character: uint32(startChar), //nolint:gosec
		},
		End: protocol.Position{
			Line:      uint32(endLine), //nolint:gosec
			Character: uint32(endChar), //nolint:gosec
		},
	}
}

// getParsedFile returns the parsed stylesheet for a URI,
// protected by a read lock.
func (h *cssHandler) getParsedFile(uri string) *parser.Stylesheet {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.parsedFiles[uri]
}

// getRawFile returns the raw source for a URI, protected by a
// read lock.
func (h *cssHandler) getRawFile(uri string) []byte {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.rawFiles[uri]
}

// --- server.Handler implementation ---

func (h *cssHandler) Initialize(
	_ context.Context,
	params *protocol.InitializeParams,
) (protocol.ServerCapabilities, error) {
	if params.InitializationOptions != nil {
		// Decode settings from initializationOptions.
		// The options arrive as map[string]any from JSON.
		if opts, ok := params.InitializationOptions.(map[string]any); ok {
			h.settings = &ServerSettings{}
			if v, ok := opts["formatMode"].(string); ok {
				h.settings.FormatMode = v
			}
			if v, ok := opts["printWidth"].(float64); ok {
				h.settings.PrintWidth = int(v)
			}
			if v, ok := opts["experimentalFeatures"].(string); ok {
				h.settings.ExperimentalFeatures = v
			}
			if v, ok := opts["deprecatedFeatures"].(string); ok {
				h.settings.DeprecatedFeatures = v
			}
			if v, ok := opts["unknownValues"].(string); ok {
				h.settings.UnknownValues = v
			}
			if v, ok := opts["strictColorNames"].(bool); ok {
				h.settings.StrictColorNames = v
			}
		}

		if h.settings != nil {
			h.lintOpts.Experimental = modeFromString(
				h.settings.ExperimentalFeatures,
				analyzer.ExperimentalIgnore,
				analyzer.ExperimentalError,
				analyzer.ExperimentalWarn,
			)
			h.lintOpts.Deprecated = modeFromString(
				h.settings.DeprecatedFeatures,
				analyzer.DeprecatedIgnore,
				analyzer.DeprecatedError,
				analyzer.DeprecatedWarn,
			)
			h.lintOpts.UnknownValues = modeFromString(
				h.settings.UnknownValues,
				analyzer.UnknownValueIgnore,
				analyzer.UnknownValueError,
				analyzer.UnknownValueWarn,
			)
			h.lintOpts.StrictColorNames = h.settings.StrictColorNames
		}
	}

	// Extract root URI and scan workspace.
	var rootURI string
	if len(params.WorkspaceFolders) > 0 {
		rootURI = params.WorkspaceFolders[0].URI
	} else if params.RootURI != "" { //nolint:staticcheck // fallback for older clients
		rootURI = string(params.RootURI) //nolint:staticcheck // fallback
	}
	if rootURI != "" {
		h.rootPath = pathutil.URIToFilePath(rootURI)
		if h.rootPath != "" {
			_ = h.varIndex.ScanWorkspace(h.rootPath)
		}
	}

	return protocol.ServerCapabilities{
		HoverProvider: true,
		CompletionProvider: &protocol.CompletionOptions{
			TriggerCharacters: []string{":", "@", ".", "#", "-", " "},
		},
		DefinitionProvider:         true,
		ReferencesProvider:         true,
		RenameProvider:             &protocol.RenameOptions{PrepareProvider: true},
		DocumentFormattingProvider: true,
		CodeActionProvider: &protocol.CodeActionOptions{
			CodeActionKinds: []protocol.CodeActionKind{
				protocol.CodeActionKind(analyzer.CodeActionQuickFix),
				protocol.CodeActionKind(analyzer.CodeActionRefactor),
				protocol.CodeActionKind(analyzer.CodeActionSourceFixAll),
			},
		},
		DocumentSymbolProvider:    true,
		ColorProvider:             true,
		DocumentHighlightProvider: true,
		FoldingRangeProvider:      true,
		DocumentLinkProvider:      &protocol.DocumentLinkOptions{},
		SelectionRangeProvider:    true,
	}, nil
}

func (h *cssHandler) Diagnostics(
	_ context.Context,
	uri protocol.DocumentURI,
	content string,
) ([]protocol.Diagnostic, error) {
	src := []byte(content)
	diags, ss := css.Diagnostics(src, h.lintOpts)

	h.mu.Lock()
	h.rawFiles[string(uri)] = src
	h.parsedFiles[string(uri)] = ss
	h.mu.Unlock()

	h.varIndex.IndexFileWithStylesheet(string(uri), ss, src)

	result := make([]protocol.Diagnostic, len(diags))
	for i, d := range diags {
		result[i] = protocol.Diagnostic{
			Range: protocol.Range{
				Start: protocol.Position{
					Line:      uint32(d.StartLine), //nolint:gosec
					Character: uint32(d.StartChar), //nolint:gosec
				},
				End: protocol.Position{
					Line:      uint32(d.EndLine), //nolint:gosec
					Character: uint32(d.EndChar), //nolint:gosec
				},
			},
			Message:  d.Message,
			Severity: protocol.DiagnosticSeverity(d.Severity),
		}
	}

	return result, nil
}

func (h *cssHandler) Shutdown(context.Context) error {
	return nil
}

// --- server.HoverHandler ---

func (h *cssHandler) Hover(
	_ context.Context,
	params *protocol.HoverParams,
) (*protocol.Hover, error) { //nolint:unparam // interface
	uri := string(params.TextDocument.URI)
	src := h.getRawFile(uri)
	if src == nil {
		return nil, nil
	}
	ss := h.getParsedFile(uri)
	if ss == nil {
		result := css.Parse(src)
		ss = result.Stylesheet
	}

	hover := css.Hover(
		ss, src,
		int(params.Position.Line),      //nolint:gosec
		int(params.Position.Character), //nolint:gosec
		h.varIndex,
	)

	if !hover.Found {
		return nil, nil
	}

	result := &protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind:  protocol.Markdown,
			Value: hover.Content,
		},
	}
	if hover.RangeStart < hover.RangeEnd {
		r := offsetRangeToProtocolRange(
			src, hover.RangeStart, hover.RangeEnd,
		)
		result.Range = &r
	}

	return result, nil
}

// --- server.CompletionHandler ---

func (h *cssHandler) Completion(
	_ context.Context,
	params *protocol.CompletionParams,
) (*protocol.CompletionList, error) { //nolint:unparam // interface
	uri := string(params.TextDocument.URI)
	src := h.getRawFile(uri)
	if src == nil {
		return nil, nil
	}
	ss := h.getParsedFile(uri)
	if ss == nil {
		result := css.Parse(src)
		ss = result.Stylesheet
	}

	items := css.Completions(
		ss, src,
		int(params.Position.Line),      //nolint:gosec
		int(params.Position.Character), //nolint:gosec
		h.lintOpts,
	)

	lspItems := make([]protocol.CompletionItem, len(items))
	for i, item := range items {
		lspItems[i] = protocol.CompletionItem{
			Label:      item.Label,
			Kind:       protocol.CompletionItemKind(item.Kind),
			Detail:     item.Detail,
			InsertText: item.InsertText,
			Deprecated: item.Deprecated,
		}
		if len(item.Tags) > 0 {
			tags := make([]protocol.CompletionItemTag, len(item.Tags))
			for j, t := range item.Tags {
				tags[j] = protocol.CompletionItemTag(t)
			}
			lspItems[i].Tags = tags
		}
		if item.Documentation != "" {
			lspItems[i].Documentation = &protocol.MarkupContent{
				Kind:  protocol.Markdown,
				Value: item.Documentation,
			}
		}
	}

	return &protocol.CompletionList{
		IsIncomplete: false,
		Items:        lspItems,
	}, nil
}

// --- server.DefinitionHandler ---

func (h *cssHandler) Definition(
	_ context.Context,
	params *protocol.DefinitionParams,
) ([]protocol.Location, error) { //nolint:unparam // interface
	uri := string(params.TextDocument.URI)
	src := h.getRawFile(uri)
	if src == nil {
		return nil, nil
	}
	ss := h.getParsedFile(uri)
	if ss == nil {
		result := css.Parse(src)
		ss = result.Stylesheet
	}

	defResult, found := css.Definition(
		ss, src,
		int(params.Position.Line),      //nolint:gosec
		int(params.Position.Character), //nolint:gosec
	)

	if found {
		targetRange := offsetRangeToProtocolRange(
			src, defResult.TargetStart, defResult.TargetEnd,
		)
		return []protocol.Location{{
			URI:   params.TextDocument.URI,
			Range: targetRange,
		}}, nil
	}

	// Fall back to workspace index for cross-file lookup.
	varName, _, _ := css.VarReferenceWithRange(
		ss, src,
		int(params.Position.Line),      //nolint:gosec
		int(params.Position.Character), //nolint:gosec
	)
	if varName == "" {
		return nil, nil
	}

	defs := h.varIndex.LookupDefinitions(varName)
	if len(defs) == 0 {
		return nil, nil
	}

	def := defs[0]
	targetSrc := h.getRawFile(def.URI)
	if targetSrc == nil {
		path := pathutil.URIToFilePath(def.URI)
		if path == "" {
			return nil, nil
		}
		var readErr error
		targetSrc, readErr = os.ReadFile(path) //nolint:gosec
		if readErr != nil {
			return nil, nil //nolint:nilerr // file unreadable, no definition
		}
	}

	targetRange := offsetRangeToProtocolRange(
		targetSrc, def.StartPos, def.EndPos,
	)
	return []protocol.Location{{
		URI:   protocol.DocumentURI(def.URI),
		Range: targetRange,
	}}, nil
}

// --- server.FormattingHandler ---

func (h *cssHandler) Formatting(
	_ context.Context,
	params *protocol.DocumentFormattingParams,
) ([]protocol.TextEdit, error) { //nolint:unparam // interface
	uri := string(params.TextDocument.URI)
	src := h.getRawFile(uri)
	if src == nil {
		return nil, nil
	}
	ss := h.getParsedFile(uri)
	if ss == nil {
		result := css.Parse(src)
		ss = result.Stylesheet
	}

	fmtOpts := analyzer.FormatOptions{
		TabSize:      int(params.Options.TabSize),
		InsertSpaces: params.Options.InsertSpaces,
	}
	if h.settings != nil {
		switch h.settings.FormatMode {
		case "compact":
			fmtOpts.Mode = analyzer.FormatCompact
		case "preserve":
			fmtOpts.Mode = analyzer.FormatPreserve
		case "detect":
			fmtOpts.Mode = analyzer.FormatDetect
		}
		fmtOpts.PrintWidth = h.settings.PrintWidth
	}

	formatted := css.FormatDocument(ss, src, fmtOpts)

	return []protocol.TextEdit{{
		Range:   offsetRangeToProtocolRange(src, 0, len(src)),
		NewText: formatted,
	}}, nil
}

// --- server.CodeActionHandler ---

func (h *cssHandler) CodeAction(
	_ context.Context,
	params *protocol.CodeActionParams,
) ([]protocol.CodeAction, error) { //nolint:unparam // interface
	uri := string(params.TextDocument.URI)
	src := h.getRawFile(uri)
	if src == nil {
		return nil, nil
	}

	// Handle source.fixAll requests.
	if containsOnly(params.Context.Only, analyzer.CodeActionSourceFixAll) {
		return h.fixAll(params, src), nil
	}

	// Convert protocol diagnostics to analyzer diagnostics.
	var analyzerDiags []analyzer.Diagnostic
	for _, d := range params.Context.Diagnostics {
		analyzerDiags = append(analyzerDiags, analyzer.Diagnostic{
			Message:   d.Message,
			StartLine: int(d.Range.Start.Line),      //nolint:gosec
			StartChar: int(d.Range.Start.Character), //nolint:gosec
			EndLine:   int(d.Range.End.Line),        //nolint:gosec
			EndChar:   int(d.Range.End.Character),   //nolint:gosec
			Severity:  int(d.Severity),
		})
	}

	ss := h.getParsedFile(uri)
	if ss == nil {
		result := css.Parse(src)
		ss = result.Stylesheet
	}

	cursorLine := int(params.Range.Start.Line)      //nolint:gosec
	cursorChar := int(params.Range.Start.Character) //nolint:gosec

	actions := css.CodeActions(
		ss, src, cursorLine, cursorChar, analyzerDiags,
	)
	result := make([]protocol.CodeAction, len(actions))
	for i, a := range actions {
		kind := protocol.CodeActionKind(a.Kind)
		result[i] = protocol.CodeAction{
			Title: a.Title,
			Kind:  kind,
			Edit: &protocol.WorkspaceEdit{
				Changes: map[protocol.DocumentURI][]protocol.TextEdit{
					protocol.DocumentURI(uri): {{
						Range: protocol.Range{
							Start: protocol.Position{
								Line:      uint32(a.StartLine), //nolint:gosec
								Character: uint32(a.StartChar), //nolint:gosec
							},
							End: protocol.Position{
								Line:      uint32(a.EndLine), //nolint:gosec
								Character: uint32(a.EndChar), //nolint:gosec
							},
						},
						NewText: a.ReplaceWith,
					}},
				},
			},
		}
	}

	return result, nil
}

// containsOnly checks if a code action kind is in the only
// filter list.
func containsOnly(only []protocol.CodeActionKind, kind string) bool {
	return slices.Contains(only, protocol.CodeActionKind(kind))
}

// fixAll computes all fixable edits and returns them as a
// single source.fixAll code action.
func (h *cssHandler) fixAll(
	params *protocol.CodeActionParams,
	src []byte,
) []protocol.CodeAction {
	actions := css.FixAllActions(src, h.lintOpts)
	if len(actions) == 0 {
		return nil
	}

	edits := make([]protocol.TextEdit, len(actions))
	for i, a := range actions {
		edits[i] = protocol.TextEdit{
			Range: protocol.Range{
				Start: protocol.Position{
					Line:      uint32(a.StartLine), //nolint:gosec
					Character: uint32(a.StartChar), //nolint:gosec
				},
				End: protocol.Position{
					Line:      uint32(a.EndLine), //nolint:gosec
					Character: uint32(a.EndChar), //nolint:gosec
				},
			},
			NewText: a.ReplaceWith,
		}
	}

	kind := protocol.CodeActionKind(analyzer.CodeActionSourceFixAll)
	return []protocol.CodeAction{{
		Title: "Fix all auto-fixable problems",
		Kind:  kind,
		Edit: &protocol.WorkspaceEdit{
			Changes: map[protocol.DocumentURI][]protocol.TextEdit{
				params.TextDocument.URI: edits,
			},
		},
	}}
}

// --- server.ReferencesHandler ---

func (h *cssHandler) References(
	_ context.Context,
	params *protocol.ReferenceParams,
) ([]protocol.Location, error) { //nolint:unparam // interface
	uri := string(params.TextDocument.URI)
	src := h.getRawFile(uri)
	if src == nil {
		return nil, nil
	}
	ss := h.getParsedFile(uri)
	if ss == nil {
		result := css.Parse(src)
		ss = result.Stylesheet
	}

	refs := css.References(
		ss, src,
		int(params.Position.Line),      //nolint:gosec
		int(params.Position.Character), //nolint:gosec
	)

	result := make([]protocol.Location, len(refs))
	for i, ref := range refs {
		result[i] = protocol.Location{
			URI: params.TextDocument.URI,
			Range: offsetRangeToProtocolRange(
				src, ref.StartPos, ref.EndPos,
			),
		}
	}

	return result, nil
}

// --- server.RenameHandler ---

func (h *cssHandler) Rename(
	_ context.Context,
	params *protocol.RenameParams,
) (*protocol.WorkspaceEdit, error) { //nolint:unparam // interface
	uri := string(params.TextDocument.URI)
	src := h.getRawFile(uri)
	if src == nil {
		return nil, nil
	}
	ss := h.getParsedFile(uri)
	if ss == nil {
		result := css.Parse(src)
		ss = result.Stylesheet
	}

	edits := css.Rename(
		ss, src,
		int(params.Position.Line),      //nolint:gosec
		int(params.Position.Character), //nolint:gosec
		params.NewName,
	)

	if len(edits) == 0 {
		return nil, nil
	}

	textEdits := make([]protocol.TextEdit, len(edits))
	for i, e := range edits {
		textEdits[i] = protocol.TextEdit{
			Range: offsetRangeToProtocolRange(
				src, e.StartPos, e.EndPos,
			),
			NewText: e.NewText,
		}
	}

	return &protocol.WorkspaceEdit{
		Changes: map[protocol.DocumentURI][]protocol.TextEdit{
			params.TextDocument.URI: textEdits,
		},
	}, nil
}

// --- server.DocumentSymbolHandler ---

func (h *cssHandler) DocumentSymbol(
	_ context.Context,
	params *protocol.DocumentSymbolParams,
) ([]any, error) { //nolint:unparam // interface
	uri := string(params.TextDocument.URI)
	src := h.getRawFile(uri)
	if src == nil {
		return nil, nil
	}
	ss := h.getParsedFile(uri)
	if ss == nil {
		result := css.Parse(src)
		ss = result.Stylesheet
	}

	symbols := css.DocumentSymbols(ss, src)
	result := convertSymbols(symbols, src)

	out := make([]any, len(result))
	for i, s := range result {
		out[i] = s
	}

	return out, nil
}

func convertSymbols(
	symbols []analyzer.DocumentSymbol,
	src []byte,
) []protocol.DocumentSymbol {
	result := make([]protocol.DocumentSymbol, len(symbols))
	for i, s := range symbols {
		result[i] = protocol.DocumentSymbol{
			Name: s.Name,
			Kind: protocol.SymbolKind(s.Kind),
			Range: offsetRangeToProtocolRange(
				src, s.StartPos, s.EndPos,
			),
			SelectionRange: offsetRangeToProtocolRange(
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

	handler := newCSSHandler()
	srv := server.Server{
		Name:    serverName,
		Version: version,
		Handler: handler,
	}

	_ = srv.Run(context.Background())
}
