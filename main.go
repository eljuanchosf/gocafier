package main

import (
	"fmt"

	"gopkg.in/gomail.v2"

	"github.com/eljuanchosf/gocafier/Godeps/_workspace/src/gopkg.in/alecthomas/kingpin.v2"
	"github.com/eljuanchosf/gocafier/caching"
	"github.com/eljuanchosf/gocafier/logging"
	"github.com/eljuanchosf/gocafier/ocaclient"
	"github.com/eljuanchosf/gocafier/settings"
	"github.com/kr/pretty"
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
	logging.LogStd(fmt.Sprintf("Starting gocafier %s ", version), true)
	logging.SetupLogging(*debug)

	kingpin.Version(version)
	kingpin.Parse()
	config = settings.LoadConfig("config.yml")
	caching.CreateBucket(*cachePath)

	packageType := "paquetes"
	packageNumber := "3867500000015544038"
	logging.LogPackage(packageNumber, "Verifying...")
	pastData, err := caching.GetPackage(packageNumber)
	if err != nil {
		panic(err)
	}

	//pastData.Data[0].Log[0] = caching.DetailLog{}
	currentData, err := ocaclient.RequestData(packageType, packageNumber)
	if err != nil {
		panic(err)
	}

	currentData.Data[0].Type = packageType

	if pastData == nil {
		logging.LogPackage(packageNumber, "Package does not exists in cache, saving initial data.")
		//currentData.Save()
		logging.LogPackage(packageNumber, "Sending notification...")
		err = SendNotification(currentData)
		if err != nil {
			panic(err)
		}
	} else {
		diff, diffFound := pastData.DiffWith(currentData)
		if diffFound {
			logging.LogPackage(packageNumber, "Change detected. Sending notification.")
			fmt.Printf("%# v", pretty.Formatter(diff))
			//currentData.Save()
		} else {
			logging.LogPackage(packageNumber, "No change.")
		}
	}
	caching.Close()
}

func SendNotification(currentData caching.OcaPackageDetail) error {
	packageCode := currentData.Data[0].Code
	m := gomail.NewMessage()
	m.SetHeader("From", config.Email.From)
	m.SetHeader("To", config.Email.To)
	m.SetHeader("Subject", fmt.Sprintf(config.Email.Subject, packageCode))
	m.SetBody("text/html", fmt.Sprintf(config.Email.Body, packageCode))

	d := gomail.NewPlainDialer(config.SMTP.Server, config.SMTP.Port, *smtpUser, *smtpPassword)
	if err := d.DialAndSend(m); err != nil {
		return err
	}
	logging.LogPackage(packageCode, "Notification sent")
	return nil
}
