package eslogrus

import (
	"fmt"
	"log"
	"testing"

	elastic "github.com/elastic/go-elasticsearch"
	"github.com/sirupsen/logrus"
)

var esClient *elastic.Client

func InitEs() {
	cfg := elastic.Config{
		Addresses: []string{
			"http://localhost:9200",
		},
	}
	client, err := elastic.NewClient(cfg)
	if err != nil {
		log.Panic(err)
	}
	esClient = client
}

func TestEsLogrus(t *testing.T) {
	InitEs()
	hook, err := NewElasticHook(esClient, "localhost", logrus.DebugLevel, "my_index")
	if err != nil {
		fmt.Println("err", err)
	}
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})
	logger.AddHook(hook)
	logger.Error("这是一个测试情况")
}
