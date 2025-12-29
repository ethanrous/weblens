package log

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/ethanrous/weblens/modules/errors"
	"github.com/opensearch-project/opensearch-go/v4"
	opensearchapi "github.com/opensearch-project/opensearch-go/v4/opensearchapi"
)

// Message represents a structured log message for OpenSearch.
type Message struct {
	Timestamp time.Time `json:"@timestamp"`
	Message   string    `json:"message"`
	Level     string    `json:"level"`
}

// NewOpenSearchClient creates a new OpenSearch client with the given credentials.
func NewOpenSearchClient(opensearchURL, username, password string) (*opensearch.Client, error) {
	cfg := opensearch.Config{
		Addresses: []string{opensearchURL},
		Username:  username,
		Password:  password,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	client, err := opensearch.NewClient(cfg)
	if err != nil {
		return nil, errors.Errorf("error creating OpenSearch client: %v", err)
	}

	settings := strings.NewReader(`{
     "settings": {
       "index": {
            "number_of_shards": 1,
            "number_of_replicas": 2
            }
          }
     }`)

	newIndexReq := opensearchapi.IndicesCreateReq{
		Index: "weblens_dev",
		Body:  settings,
	}

	res, err := client.Do(context.Background(), newIndexReq, nil)
	if err != nil {
		return nil, errors.Errorf("error calling opensearch creating index: %v", err)
	}

	defer res.Body.Close() //nolint:errcheck

	var responseData map[string]any

	if res.StatusCode >= http.StatusBadRequest {
		bodyBytes, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, errors.Errorf("error reading response body: %v", err)
		}

		err = json.Unmarshal(bodyBytes, &responseData)
		if err != nil {
			return nil, errors.Errorf("error unmarshaling create index body: %v", err)
		}

		if res.StatusCode == http.StatusBadRequest && responseData["error"].(map[string]any)["type"] == "resource_already_exists_exception" {
			return client, nil
		}

		fmt.Printf("REASON: %v\n", responseData)

		return nil, errors.Errorf("opensearch error creating index: %s", res.String())
	}

	return client, nil
}

// NewLogMessage creates a new log message with the current timestamp.
func NewLogMessage(msg, level string) Message {
	return Message{
		Timestamp: time.Now().UTC(),
		Message:   msg,
		Level:     level,
	}
}

// OpensearchLogger is a logger that writes to OpenSearch.
type OpensearchLogger struct {
	client    *opensearch.Client
	msgQueue  chan []byte
	indexName string
}

// NewOpensearchLogger creates a new OpenSearch logger with worker goroutines.
func NewOpensearchLogger(client *opensearch.Client, indexName string) *OpensearchLogger {
	l := &OpensearchLogger{
		client:    client,
		indexName: indexName,
		msgQueue:  make(chan []byte, 100),
	}

	var loggerWorkerCount = runtime.NumCPU() / 2
	for range loggerWorkerCount {
		go l.worker()
	}

	return l
}

func (l *OpensearchLogger) Write(msg []byte) (int, error) {
	msgCpy := make([]byte, len(msg))
	copy(msgCpy, msg)

	l.msgQueue <- msgCpy

	return len(msg), nil
}

func (l *OpensearchLogger) worker() {
	for msg := range l.msgQueue {
		err := l.send(msg)
		if err != nil {
			NewZeroLogger(CreateOpts{NoOpenSearch: true}).Error().Stack().Err(err).Msgf("failed sending log message to OpenSearch: %s", string(msg))
		}
	}
}

func (l *OpensearchLogger) send(msg []byte) error {
	var logMsg map[string]any

	// Remove duplicate keys
	err := json.Unmarshal(msg, &logMsg)
	if err != nil {
		return err
	}

	msg, err = json.Marshal(logMsg)
	if err != nil {
		return err
	}

	req := opensearchapi.IndexReq{
		Index: l.indexName,
		Body:  bytes.NewReader(msg),
	}

	resp, err := l.client.Do(context.Background(), req, nil)
	if err != nil {
		return errors.Errorf("error sending request to OpenSearch: %v", err)
	}

	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode >= 300 {
		return errors.Errorf("error response from OpenSearch: %s", resp.String())
	}

	return nil
}
