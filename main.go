package main

import (
	"fmt"

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

	for _, packageNumber := range settings.Values.Packages {
		log.LogPackage(packageNumber, "Verifying...")
		pastData, err := caching.GetPackage(packageNumber)
		if err != nil {
			panic(err)
		}

		packageType := "paquetes"
		//pastData.Data[0].Log[0] = caching.DetailLog{}
		currentData, err := ocaclient.RequestData(packageType, packageNumber)
		if err != nil {
			panic(err)
		}

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
	}
	caching.Close()
}

func changeDetected(packageNumber string, currentData caching.OcaPackageDetail, diff []caching.DetailLog) {
	var err error
	log.LogPackage(packageNumber, "Change detected.")
	err = notifications.Send(currentData, diff, *smtpUser, *smtpPassword)
	if err != nil {
		panic(err)
	}
	//currentData.Save()
}
