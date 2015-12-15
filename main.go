package main

import (
	"fmt"

	"github.com/eljuanchosf/gocafier/Godeps/_workspace/src/gopkg.in/alecthomas/kingpin.v2"
	"github.com/eljuanchosf/gocafier/caching"
	"github.com/eljuanchosf/gocafier/logging"
	"github.com/eljuanchosf/gocafier/ocaclient"
	"github.com/kr/pretty"
)

var (
	debug      = kingpin.Flag("debug", "Enable debug mode. This disables emailing").Default("false").OverrideDefaultFromEnvar("GOCAFIER_DEBUG").Bool()
	cachePath  = kingpin.Flag("cache-path", "Bolt Database path ").Default("").OverrideDefaultFromEnvar("GOCAFIER_CACHE_PATH").String()
	tickerTime = kingpin.Flag("ticker-time", "Poller interval in secs").Default("3600s").OverrideDefaultFromEnvar("GOCAFIER_PULL_TIME").Duration()
	configPath = kingpin.Flag("config-path", "Set the Path to write profiling file").Default(".").OverrideDefaultFromEnvar("GOCAFIER_PATH_PROF").String()
)

const (
	version = "1.0.0"
)

func main() {
	logging.LogStd(fmt.Sprintf("Starting gocafier %s ", version), true)
	logging.SetupLogging(*debug)

	kingpin.Version(version)
	kingpin.Parse()
	caching.CreateBucket(*cachePath)

	packageType := "paquetes"
	packageNumber := "3867500000015544782"

	ocaData, err := ocaclient.RequestData(packageType, packageNumber)
	if err != nil {
		panic(err)
	}
	ocaData.Save()
	savedPackage, err := caching.GetPackage(packageNumber)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%# v", pretty.Formatter(savedPackage))
	caching.Close()
}
