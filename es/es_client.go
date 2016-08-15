package es

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"gopkg.in/olivere/elastic.v3"

	"github.com/uber-go/zap"
)

type ESClient struct {
	Debug  bool
	Host   string
	Port   int
	Index  string
	Logger zap.Logger
	Sniff  bool
	Client *elastic.Client
}

var once sync.Once
var client *ESClient

func GetESClient(host string, port int, index string, sniff bool, logger zap.Logger, debug bool) *ESClient {
	once.Do(func() {
		client = &ESClient{
			Debug:  debug,
			Host:   host,
			Port:   port,
			Logger: logger,
			Index:  index,
			Sniff:  sniff,
		}
		client.configure()
	})
	return client
}

func (es *ESClient) configure() {
	es.configureESClient()
}

func (es *ESClient) configureESClient() {
	l := es.Logger.With(
		zap.String("source", "elasticsearch"),
		zap.String("operation", "configureEsClient"),
	)
	l.Info("Connecting to elasticsearch...", zap.String("elasticsearch.url", fmt.Sprintf("http://%s:%d/%s", es.Host, es.Port, es.Index)), zap.Bool("sniff", es.Sniff))
	var err error
	es.Client, err = elastic.NewClient(
		elastic.SetURL(fmt.Sprintf("http://%s:%d", es.Host, es.Port)),
		elastic.SetSniff(es.Sniff),
	)
	if err != nil {
		l.Error("Failed to connect to elasticsearch!", zap.String("elasticsearch.url", fmt.Sprintf("http://%s:%d/%s", es.Host, es.Port, es.Index)), zap.Error(err))
		os.Exit(1)
	}
	client := es.Client
	_, err = client.CreateIndex(es.Index).Do()
	if err != nil {
		if strings.Contains(err.Error(), "index_already_exists_exception") || strings.Contains(err.Error(), "IndexAlreadyExistsException") {
			l.Warn("Index already exists into ES! Ignoring creation...", zap.String("index", es.Index))
		} else {
			l.Error("Failed to create index into ES", zap.Error(err))
			os.Exit(1)
		}
	} else {
		l.Info("Sucessfully created index into ES", zap.String("index", es.Index))
	}
}
