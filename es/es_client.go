package es

import (
	"fmt"
	"os"
	"sync"

	"gopkg.in/olivere/elastic.v3"

	"github.com/topfreegames/khan/log"
	"github.com/uber-go/zap"
)

// Client is the struct of an elasticsearch client
type Client struct {
	Debug  bool
	Host   string
	Port   int
	Index  string
	Logger zap.Logger
	Sniff  bool
	Client *elastic.Client
}

var once sync.Once
var client *Client

// GetIndexName returns the name of the index
func (client *Client) GetIndexName(gameID string) string {
	if client.Index != "" {
		return fmt.Sprintf("%s-%s", client.Index, gameID)
	}
	return "khan-test"
}

// GetClient returns an elasticsearch client configures with the given the arguments
func GetClient(host string, port int, index string, sniff bool, logger zap.Logger, debug bool) *Client {
	once.Do(func() {
		client = &Client{
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

// GetTestClient returns a test elasticsearch client configures with the given the arguments
func GetTestClient(host string, port int, index string, sniff bool, logger zap.Logger, debug bool) *Client {
	client = &Client{
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

// GetConfiguredClient returns an elasticsearch client with no extra configs
func GetConfiguredClient() *Client {
	return client
}

func (es *Client) configure() {
	es.configureClient()
}

// DestroyClient sets the elasticsearch client value to nil
func DestroyClient() {
	client = nil
}

func (es *Client) configureClient() {
	l := es.Logger.With(
		zap.String("source", "elasticsearch"),
		zap.String("operation", "configureClient"),
	)
	log.I(l, "Connecting to elasticsearch...", func(cm log.CM) {
		cm.Write(
			zap.String("elasticsearch.url", fmt.Sprintf("http://%s:%d/%s", es.Host, es.Port, es.Index)),
			zap.Bool("sniff", es.Sniff),
		)
	})
	var err error
	es.Client, err = elastic.NewClient(
		elastic.SetURL(fmt.Sprintf("http://%s:%d", es.Host, es.Port)),
		elastic.SetSniff(es.Sniff),
	)
	if err != nil {
		log.E(l, "Failed to connect to elasticsearch!", func(cm log.CM) {
			cm.Write(
				zap.String("elasticsearch.url", fmt.Sprintf("http://%s:%d/%s", es.Host, es.Port, es.Index)),
				zap.Error(err),
			)
		})
		os.Exit(1)
	}
}
