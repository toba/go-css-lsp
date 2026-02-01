package lsp

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"log"
	"strconv"
)

// ReceiveInput creates a scanner that decodes LSP messages from
// an input stream.
func ReceiveInput(input io.Reader) *bufio.Scanner {
	scanner := bufio.NewScanner(input)
	scanner.Split(decode)
	return scanner
}

// SendOutput writes raw bytes to the output writer.
func SendOutput(output io.Writer, response []byte) {
	_, err := output.Write(response)
	if err != nil {
		log.Printf(
			"Error while writing to output: %s",
			err.Error(),
		)
	}
}

// SendToLspClient sends an LSP response to the client.
func SendToLspClient(output io.Writer, response []byte) {
	response = Encode(response)
	SendOutput(output, response)
}

// Encode wraps data with Content-Length header per LSP spec.
func Encode(dataContent []byte) []byte {
	length := strconv.Itoa(len(dataContent))
	dataHeader := []byte(
		ContentLengthHeader + ": " + length + HeaderDelimiter,
	)
	dataHeader = append(dataHeader, dataContent...)
	return dataHeader
}

// decode is a bufio.SplitFunc that parses LSP messages.
func decode(
	data []byte,
	_ bool,
) (advance int, token []byte, err error) {
	indexStartData := bytes.Index(data, []byte(HeaderDelimiter))
	if indexStartData == -1 {
		return 0, nil, nil
	}

	contentLength, parseErr := getHeaderContentLength(
		data[:indexStartData],
	)
	if parseErr != nil {
		return indexStartData + 4, []byte{}, parseErr
	}

	if len(data[indexStartData:]) < contentLength {
		return 0, nil, nil
	}

	indexStartData += 4
	indexEndData := indexStartData + contentLength

	return indexEndData, data[indexStartData:indexEndData], nil
}

// getHeaderContentLength extracts Content-Length from LSP headers.
func getHeaderContentLength(data []byte) (int, error) {
	indexHeader := bytes.LastIndex(
		data,
		[]byte(ContentLengthHeader),
	)
	if indexHeader == -1 {
		return -1, errors.New(
			"unable to find '" + ContentLengthHeader + "' header",
		)
	}

	indexLineSeparator := bytes.Index(
		data[indexHeader:],
		[]byte(LineDelimiter),
	)
	if indexLineSeparator >= 0 {
		indexLineSeparator += indexHeader
	} else {
		indexLineSeparator = len(data)
	}

	indexKeyValueSeparator := bytes.Index(
		data[indexHeader:indexLineSeparator], []byte(":"),
	)
	if indexKeyValueSeparator == -1 {
		return -1, errors.New(
			"malformed '" + ContentLengthHeader + "' header: missing ':'",
		)
	}

	indexKeyValueSeparator += indexHeader

	contentLengthString := data[indexKeyValueSeparator+1 : indexLineSeparator]
	contentLengthString = bytes.TrimSpace(contentLengthString)

	contentLength, err := strconv.Atoi(
		string(contentLengthString),
	)
	if err != nil {
		return -1, errors.New(
			"malformed '" + ContentLengthHeader + "' header: value is not an integer",
		)
	}

	if contentLength < 0 {
		return -1, errors.New(
			"'" + ContentLengthHeader + "' cannot be negative",
		)
	}

	return contentLength, nil
}
