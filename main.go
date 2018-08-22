package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/alileza/frog/config"
	"github.com/alileza/frog/consumer"
	"github.com/alileza/frog/evaluator"
	"github.com/alileza/frog/storage"
	"github.com/alileza/frog/util/version"
	"github.com/pkg/errors"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	configFile  string
	storagePath string
	listenAddr  string
)

func main() {
	app := kingpin.New(filepath.Base(os.Args[0]), "frog - message queue diff and docs generator")
	app.Version(version.Print())
	app.HelpFlag.Short('h')

	app.Flag("config.file", "frog configuration file path.").Short('c').Default("frog.yml").StringVar(&configFile)
	app.Flag("storage.path", "report storage path").Short('o').Default("reports").StringVar(&storagePath)
	app.Flag("http.listen-address", "report storage path").Short('p').Default("0.0.0.0:9000").StringVar(&listenAddr)

	_, err := app.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "Error parsing flag"))
		os.Exit(1)
	}

	cfg, err := config.Retrieve(configFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "Error retrieving config"))
		os.Exit(1)
	}

	consumerClient := consumer.New(&consumer.Options{
		DSN:     cfg.DSN,
		Targets: cfg.Targets,
	})
	defer func() {
		consumerClient.Close()
	}()

	go consumerClient.Run()

	evaluatorClient := evaluator.New()
	storageClient := storage.New(storagePath)

	go func() {
		for {
			msg := <-consumerClient.ListenMessage()
			report := evaluatorClient.Eval(msg.Target, msg.Body)
			if report == nil {
				continue
			}
			if err := storageClient.Store(report.Name, report.Schema, report.Body, report.Error); err != nil {
				log.Printf("%+v : %v", report, err)
			}
		}
	}()

	webErrChan := make(chan error)
	go func(errChan chan error) {
		mux := http.NewServeMux()
		mux.Handle("/", storageClient.Handler())
		mux.Handle("/active", consumerClient.Handler())
		mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			exchange := r.FormValue("exchange")
			routingKey := r.FormValue("routingKey")
			if exchange == "" {
				fmt.Fprintf(w, `exchange is required`)
				return
			}
			if routingKey == "" {
				fmt.Fprintf(w, `routingKey is required`)
				return
			}
			if err := consumerClient.Consume(exchange + ":" + routingKey); err != nil {
				fmt.Fprintf(w, err.Error())
				return
			}
			fmt.Fprintf(w, `Registered please check your message <a href="/">here</a>`)
		})
		errChan <- http.ListenAndServe(listenAddr, mux)
	}(webErrChan)

	term := make(chan os.Signal)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)
	select {
	case <-term:
		log.Println("Received SIGTERM, exiting gracefully...")
	case err := <-consumerClient.ListenError():
		log.Println(err)
	case err := <-webErrChan:
		log.Println(err)
	}
	log.Println("tschüss 👋")
}
