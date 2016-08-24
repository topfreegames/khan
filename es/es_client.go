package es

import (
	"fmt"
	"os"
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

func GetTestESClient(host string, port int, index string, sniff bool, logger zap.Logger, debug bool) *ESClient {
	client = &ESClient{
		Debug:  debug,
		Host:   host,
		Port:   port,
		Logger: logger,
		Index:  index,
		Sniff:  sniff,
	}
	client.configure()
	return client
}

func GetConfiguredESClient() *ESClient {
	return client
}

func (es *ESClient) configure() {
	es.configureESClient()
}

func DestroyClient() {
	client = nil
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
}
