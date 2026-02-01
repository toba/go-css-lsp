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
	"sync"

	"github.com/toba/go-css-lsp/cmd/go-css-lsp/lsp"
	"github.com/toba/go-css-lsp/internal/css"
	"github.com/toba/go-css-lsp/internal/css/analyzer"
	"github.com/toba/go-css-lsp/internal/css/parser"
)

// version is set by goreleaser at build time.
var version = "dev"

// workspaceStore holds the state for a workspace.
type workspaceStore struct {
	RootPath    string
	RawFiles    map[string][]byte
	ParsedFiles map[string]*parser.Stylesheet
}

const serverName = "go-css-lsp"

// TargetFileExtensions lists supported file extensions.
var TargetFileExtensions = []string{"css"}

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
			response = lsp.ProcessIllegalRequestAfterShutdown(
				request.JsonRpc, request.Id,
			)
			lsp.SendToLspClient(os.Stdout, response)
			continue
		}

		slog.Info("request " + request.Method)

		switch request.Method {
		case lsp.MethodInitialize:
			var rootURI string
			response, rootURI = lsp.ProcessInitializeRequest(
				data, serverName, version,
			)
			notifyTheRootPath(
				rootPathNotification, rootURI,
			)
			rootPathNotification = nil
			isRequestResponse = true

		case lsp.MethodInitialized:
			isRequestResponse = false
			lsp.ProcessInitializedNotification(data)

		case lsp.MethodShutdown:
			isExiting = true
			isRequestResponse = true
			response = lsp.ProcessShutdownRequest(
				request.JsonRpc, request.Id,
			)

		case lsp.MethodDidOpen:
			isRequestResponse = false
			fileURI, fileContent :=
				lsp.ProcessDidOpenTextDocumentNotification(
					data,
				)
			insertTextDocumentToDiagnostic(
				fileURI,
				fileContent,
				textChangedNotification,
				textFromClient,
				muTextFromClient,
			)

		case lsp.MethodDidChange:
			isRequestResponse = false
			fileURI, fileContent :=
				lsp.ProcessDidChangeTextDocumentNotification(
					data,
				)
			insertTextDocumentToDiagnostic(
				fileURI,
				fileContent,
				textChangedNotification,
				textFromClient,
				muTextFromClient,
			)

		case lsp.MethodDidClose:
			isRequestResponse = false
			lsp.ProcessDidCloseTextDocumentNotification(data)

		case lsp.MethodHover:
			isRequestResponse = true
			response = processHover(
				data, storage, textFromClient,
				muTextFromClient,
			)

		case lsp.MethodCompletion:
			isRequestResponse = true
			response = processCompletion(
				data, storage, textFromClient,
				muTextFromClient,
			)

		default:
			isRequestResponse = false
		}

		if isRequestResponse {
			lsp.SendToLspClient(os.Stdout, response)
		}

		response = nil
	}

	if scanner.Err() != nil {
		msg := "error while closing LSP: " +
			scanner.Err().Error()
		slog.Error(msg)
		panic(msg)
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

	ss := storage.ParsedFiles[uri]
	if ss == nil {
		result := css.Parse(src)
		ss = result.Stylesheet
	}

	content, found := css.Hover(
		ss, src,
		int(req.Params.Position.Line),      //nolint:gosec // LSP positions are small
		int(req.Params.Position.Character), //nolint:gosec // LSP positions are small
	)

	type HoverResult struct {
		Contents lsp.MarkupContent `json:"contents"`
	}

	res := lsp.ResponseMessage[*HoverResult]{
		JsonRpc: req.JsonRpc,
		Id:      req.Id,
	}

	if found {
		res.Result = &HoverResult{
			Contents: lsp.MarkupContent{
				Kind:  "markdown",
				Value: content,
			},
		}
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

	ss := storage.ParsedFiles[uri]
	if ss == nil {
		result := css.Parse(src)
		ss = result.Stylesheet
	}

	items := css.Completions(
		ss, src,
		int(req.Params.Position.Line),      //nolint:gosec // LSP positions are small
		int(req.Params.Position.Character), //nolint:gosec // LSP positions are small
	)

	lspItems := make([]lsp.CompletionItem, len(items))
	for i, item := range items {
		lspItems[i] = lsp.CompletionItem{
			Label:      item.Label,
			Kind:       item.Kind,
			Detail:     item.Detail,
			InsertText: item.InsertText,
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
	rootURI, ok := <-rootPathNotification
	if !ok {
		return
	}

	storage.RootPath = uriToFilePath(rootURI)

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
			storage.RawFiles[uri] = content

			diags, ss := css.Diagnostics(content)
			storage.ParsedFiles[uri] = ss

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
