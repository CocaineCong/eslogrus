package eslogrus

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	elastic "github.com/elastic/go-elasticsearch"
	"github.com/elastic/go-elasticsearch/esapi"
	"github.com/sirupsen/logrus"
)

// ElasticHook is a logrus
// hook for ElasticSearch
type ElasticHook struct {
	client    *elastic.Client
	host      string
	index     IndexNameFunc
	levels    []logrus.Level
	ctx       context.Context
	ctxCancel context.CancelFunc
	fireFunc  fireFunc
}

type message struct {
	Host      string
	Timestamp string `json:"@timestamp"`
	Message   string
	Data      logrus.Fields
	Level     string
}

// IndexNameFunc get index name
type IndexNameFunc func() string

type fireFunc func(entry *logrus.Entry, hook *ElasticHook) error

// NewElasticHook
// client is an elastic client
// host is the elastic's host
// level is the level of logrus
// index is the es index you wanna create message in. by the way , if the index don't exist, the esapi will create index automatically
func NewElasticHook(client *elastic.Client, host string, level logrus.Level, index string) (*ElasticHook, error) {
	return NewElasticHookWithFunc(client, host, level, func() string { return index })
}

func NewElasticHookWithFunc(client *elastic.Client, host string, level logrus.Level, indexFunc IndexNameFunc) (*ElasticHook, error) {
	return newHookFuncAndFireFunc(client, host, level, indexFunc, syncFireFunc)
}

func newHookFuncAndFireFunc(client *elastic.Client, host string, level logrus.Level, indexFunc IndexNameFunc, fireFunc fireFunc) (*ElasticHook, error) {
	var levels []logrus.Level
	for _, l := range []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
	} {
		if l <= level {
			levels = append(levels, l)
		}
	}

	ctx, cancel := context.WithCancel(context.TODO())

	return &ElasticHook{
		client:    client,
		host:      host,
		index:     indexFunc,
		levels:    levels,
		ctx:       ctx,
		ctxCancel: cancel,
		fireFunc:  fireFunc,
	}, nil
}

func createMessage(entry *logrus.Entry, hook *ElasticHook) *message {
	level := entry.Level.String()

	if e, ok := entry.Data[logrus.ErrorKey]; ok && e != nil {
		if err, ok := e.(error); ok {
			entry.Data[logrus.ErrorKey] = err.Error()
		}
	}

	return &message{
		hook.host,
		entry.Time.UTC().Format(time.RFC3339Nano),
		entry.Message,
		entry.Data,
		strings.ToUpper(level),
	}
}

func syncFireFunc(entry *logrus.Entry, hook *ElasticHook) error {
	data, err := json.Marshal(createMessage(entry, hook))

	req := esapi.IndexRequest{
		Index:   hook.index(),
		Body:    bytes.NewReader(data),
		Refresh: "true",
	}

	res, err := req.Do(hook.ctx, hook.client)
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	var r map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		log.Printf("Error parsing the response body: %s", err)
	} else {
		// Print the response status and indexed document version.
		log.Printf("[%s] %s; version=%d", res.Status(), r["result"], int(r["_version"].(float64)))
	}
	return err
}

func (hook *ElasticHook) Fire(entry *logrus.Entry) error {
	return hook.fireFunc(entry, hook)
}

// Levels Required for logrus hook implementation
func (hook *ElasticHook) Levels() []logrus.Level {
	return hook.levels
}
