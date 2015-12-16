package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eljuanchosf/gocafier/Godeps/_workspace/src/gopkg.in/alecthomas/kingpin.v2"
	"github.com/eljuanchosf/gocafier/caching"
	log "github.com/eljuanchosf/gocafier/logging"
	"github.com/eljuanchosf/gocafier/notifications"
	"github.com/eljuanchosf/gocafier/ocaclient"
	"github.com/eljuanchosf/gocafier/settings"
)

var (
	debug        = kingpin.Flag("debug", "Enable debug mode. This disables emailing").Default("false").OverrideDefaultFromEnvar("GOCAFIER_DEBUG").Bool()
	cachePath    = kingpin.Flag("cache-path", "Bolt Database path ").Default("").OverrideDefaultFromEnvar("GOCAFIER_CACHE_PATH").String()
	tickerTime   = kingpin.Flag("ticker-time", "Poller interval in secs").Default("3600s").OverrideDefaultFromEnvar("GOCAFIER_PULL_TIME").Duration()
	configPath   = kingpin.Flag("config-path", "Set the Path to write profiling file").Default(".").OverrideDefaultFromEnvar("GOCAFIER_PATH_PROF").String()
	smtpUser     = kingpin.Flag("smtp-user", "Sets the SMTP username").Required().OverrideDefaultFromEnvar("GOCAFIER_SMTP_USER").String()
	smtpPassword = kingpin.Flag("smtp-pass", "Sets the SMTP password").Required().OverrideDefaultFromEnvar("GOCAFIER_SMTP_PASSWORD").String()
)

var config settings.Config
var ocaPackageTypes = []string{"paquetes", "cartas", "dni", "partidas"}

const (
	version = "1.0.0"
)

func main() {
	log.LogStd(fmt.Sprintf("Starting gocafier %s ", version), true)
	log.SetupLogging(*debug)

	kingpin.Version(version)
	kingpin.Parse()
	settings.LoadConfig(*configPath)
	caching.CreateBucket(*cachePath)

	log.LogStd(fmt.Sprintf("Start polling each %s", *tickerTime), true)

	//Control signal interruptions
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	go func() {
		sig := <-sigc
		switch sig {
		case os.Interrupt:
			log.LogStd("Interrupted by OS, exiting.", true)
			caching.Close()
			os.Exit(0)
		case syscall.SIGTERM:
			log.LogStd("Interrupted by SIGTERM, exiting.", true)
			caching.Close()
			os.Exit(0)
		}
	}()

	for {
		for _, packageNumber := range settings.Values.Packages {
			pastData, err := caching.GetPackage(packageNumber)
			if err != nil {
				panic(err)
			}

			currentData, packageType, packageFound := findPackage(packageNumber, pastData)

			if packageFound {
				currentData.Data[0].Type = packageType
				if pastData == nil {
					log.LogPackage(packageNumber, "Package does not exist in cache, saving initial data.")
					changeDetected(packageNumber, currentData, nil)
				} else {
					diff, diffFound := pastData.DiffWith(currentData)
					if diffFound {
						changeDetected(packageNumber, currentData, diff)
					} else {
						log.LogPackage(packageNumber, "No change.")
					}
				}
			} else {
				log.LogPackage(packageNumber, "Not found in server")
			}
		}
		time.Sleep(*tickerTime)
	}
}

func findPackage(packageNumber string, pastData *caching.OcaPackageDetail) (details caching.OcaPackageDetail, packageType string, found bool) {
	var err error
	var success bool
	found = false
	if pastData == nil {
		for _, packageType = range ocaPackageTypes {
			log.LogPackage(packageNumber, fmt.Sprintf("Checking in type '%s'", packageType))
			details, success, err = ocaclient.RequestData(packageType, packageNumber)
			if err != nil {
				panic(err)
			}
			if success {
				found = true
				break
			}
		}
	} else {
		packageType = pastData.Data[0].Type
		log.LogPackage(packageNumber, fmt.Sprintf("Found in type '%s'", packageType))
		details, success, err = ocaclient.RequestData(packageType, packageNumber)
		if err != nil {
			panic(err)
		}
		found = true
	}
	return details, packageType, found
}

func changeDetected(packageNumber string, currentData caching.OcaPackageDetail, diff []caching.DetailLog) {
	var err error
	log.LogPackage(packageNumber, "Change detected.")
	err = notifications.Send(currentData, diff, *smtpUser, *smtpPassword)
	if err != nil {
		panic(err)
	}
	currentData.Save()
}
