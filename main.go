package main

import (
	"path"
	"runtime"

	"github.com/eljuanchosf/gocafier/caching"
	"github.com/eljuanchosf/gocafier/ocaclient"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	debug      = kingpin.Flag("debug", "Enable debug mode. This disables emailing").Default("false").OverrideDefaultFromEnvar("GOCAFIER_DEBUG").Bool()
	cachePath  = kingpin.Flag("cache-path", "Bolt Database path ").Default("my.db").OverrideDefaultFromEnvar("GOCAFIER_CACHE_PATH").String()
	tickerTime = kingpin.Flag("ticker-time", "Poller interval in secs").Default("3600s").OverrideDefaultFromEnvar("GOCAFIER_PULL_TIME").Duration()
	configPath = kingpin.Flag("config-path", "Set the Path to write profiling file").Default(".").OverrideDefaultFromEnvar("GOCAFIER_PATH_PROF").String()
)

const (
	version = "1.0.0"
)

func main() {
	kingpin.Version(version)
	kingpin.Parse()

	_, filename, _, _ := runtime.Caller(0) // get full path of this file
	cacheFilename := path.Join(path.Dir(filename), *)

	//Use bolt for in-memory data caching
	caching.Open(*cachePath)
	caching.CreateBucket()

	ocaData, err := ocaclient.RequestData("paquetes", "3867500000015544782")
	if err != nil {
		panic(err)
	}

	ocaData.Save()
	caching.ListPackages()

	//fmt.Printf("%# v", pretty.Formatter(ocaData))

	caching.Close()

}
