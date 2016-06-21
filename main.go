package main

import (
	"crypto/tls"
	"fmt"
	"github.com/Financial-Times/http-handlers-go/httphandlers"
	"github.com/Financial-Times/tme-reader/tmereader"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/jawher/mow.cli"
	"github.com/rcrowley/go-metrics"
	"github.com/sethgrid/pester"
	"net"
	"net/http"
	"os"
	"time"
)

func init() {
	log.SetFormatter(new(log.JSONFormatter))
}

func main() {
	app := cli.App("v1-orgs-transformer", "A RESTful API for transforming TME Oranisations to UP json")
	username := app.String(cli.StringOpt{
		Name:   "tme-username",
		Value:  "",
		Desc:   "TME username used for http basic authentication",
		EnvVar: "TME_USERNAME",
	})
	password := app.String(cli.StringOpt{
		Name:   "tme-password",
		Value:  "",
		Desc:   "TME password used for http basic authentication",
		EnvVar: "TME_PASSWORD",
	})
	token := app.String(cli.StringOpt{
		Name:   "token",
		Value:  "",
		Desc:   "Token to be used for accessig TME",
		EnvVar: "TOKEN",
	})
	baseURL := app.String(cli.StringOpt{
		Name:   "base-url",
		Value:  "http://localhost:8080/transformers/organisations/",
		Desc:   "Base url",
		EnvVar: "BASE_URL",
	})
	tmeBaseURL := app.String(cli.StringOpt{
		Name:   "tme-base-url",
		Value:  "https://tme.ft.com",
		Desc:   "TME base url",
		EnvVar: "TME_BASE_URL",
	})
	port := app.Int(cli.IntOpt{
		Name:   "port",
		Value:  8080,
		Desc:   "Port to listen on",
		EnvVar: "PORT",
	})
	maxRecords := app.Int(cli.IntOpt{
		Name:   "maxRecords",
		Value:  int(10000),
		Desc:   "Maximum records to be queried to TME",
		EnvVar: "MAX_RECORDS",
	})
	batchSize := app.Int(cli.IntOpt{
		Name:   "batchSize",
		Value:  int(10),
		Desc:   "Number of requests to be executed in parallel to TME",
		EnvVar: "BATCH_SIZE",
	})
	cacheFileName := app.String(cli.StringOpt{
		Name:   "cache-file-name",
		Value:  "cache.db",
		Desc:   "Cache file name",
		EnvVar: "CACHE_FILE_NAME",
	})

	tmeTaxonomyName := "ON"

	app.Action = func() {
		client := getResilientClient()
		modelTransformer := new(orgTransformer)
		s := newOrgService(
			tmereader.NewTmeRepository(
				client,
				*tmeBaseURL,
				*username,
				*password,
				*token,
				*maxRecords,
				*batchSize,
				tmeTaxonomyName,
				modelTransformer),
			*baseURL,
			tmeTaxonomyName,
			*maxRecords,
			*cacheFileName)
		defer s.shutdown()
		h := newOrgsHandler(s)
		m := mux.NewRouter()
		m.HandleFunc("/transformers/organisations", h.getOrgs).Methods("GET")
		m.HandleFunc("/transformers/organisations/{uuid}", h.getOrgByUUID).Methods("GET")
		http.Handle("/", m)

		log.Printf("listening on %d", *port)
		err := http.ListenAndServe(fmt.Sprintf(":%d", *port),
			httphandlers.HTTPMetricsHandler(metrics.DefaultRegistry,
				httphandlers.TransactionAwareRequestLoggingHandler(log.StandardLogger(), m)))
		if err != nil {
			log.Errorf("Error by listen and serve: %v", err.Error())
		}
	}
	app.Run(os.Args)
}

func getResilientClient() *pester.Client {
	tr := &http.Transport{
		MaxIdleConnsPerHost: 32,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
	}
	c := &http.Client{
		Transport: tr,
		Timeout:   30 * time.Second,
	}
	client := pester.NewExtendedClient(c)
	client.Backoff = pester.ExponentialBackoff
	client.MaxRetries = 5
	client.Concurrency = 1

	return client
}
