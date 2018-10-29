package main

import (
	"fmt"
	"github.com/barskern/paragliding/igcserver"
	"github.com/globalsign/mgo"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
)

func main() {
	for _, v := range os.Args {
		switch v {
		case "-v":
			log.SetLevel(log.DebugLevel)
		case "-q":
			log.SetLevel(log.WarnLevel)
		case "-h":
			fmt.Println("Usage: paragliding [-q][-v][-h]\n\n-q Quiet mode (only warn and error)\n-v Verbose mode (all logs)")
			os.Exit(0)
		}
	}

	// Get port from env
	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "8080"
	}
	// Get mongodb url from env
	mongoURL, ok := os.LookupEnv("MONGODB_URI")
	if !ok {
		log.Fatal("unable to get required envvar 'MONGODB_URI'")
	}

	log.WithFields(log.Fields{
		"port":     port,
		"logLevel": log.GetLevel(),
	}).Info("initializing server")

	mongoSession, err := mgo.Dial(mongoURL)
	if err != nil {
		log.WithFields(log.Fields{
			"uri":   mongoURL,
			"error": err,
		}).Fatal("unable to connect to mongo db")
	}

	// Make a http client which the server will use for external requests
	httpClient := http.Client{}

	// Create a track metas abstraction which will connect to mongodb to store
	// all igctracks
	trackMetas := igcserver.NewTrackMetasDB(mongoSession.Copy())

	// Create a webhooks abstraction which will connect to a mongodb to store
	// all webhooks
	webhooks := igcserver.NewWebhooksDB(mongoSession.Copy(), &httpClient)

	// Make simple ticker for database
	ticker := igcserver.NewTickerDB(mongoSession.Copy(), 10)

	// Create a new server which encompasses all routing and server state
	server := igcserver.NewServer(&httpClient, &trackMetas, &ticker, &webhooks)

	// Route all requests to `paragliding/api/` to the server and remove prefix
	http.Handle("/paragliding/api/", http.StripPrefix("/paragliding/api", &server))
	http.Handle("/paragliding", http.RedirectHandler("/paragliding/api/", http.StatusMovedPermanently))

	// This function will block the current thread
	err = http.ListenAndServe(":"+port, nil)

	mongoSession.Close()

	// We will only get to this statement if the server unexpectedly crashes
	log.WithFields(log.Fields{
		"error": err,
	}).Fatal("server error occurred")
}
