package caching

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/boltdb/bolt"
	log "github.com/eljuanchosf/gocafier/logging"
	"github.com/kr/pretty"
	"github.com/mitchellh/go-homedir"
)

const (
	bucketName = "packages"
)

var appdb *bolt.DB
var open bool

//DetailLog represents a log of package movements
type DetailLog struct {
	Date        string `json:"date"`
	Description string `json:"description"`
}

// OcaPackageDetail represents the response from the OCA web service
type OcaPackageDetail struct {
	Data []struct {
		Type   string `json:"type"`
		Code   string `json:"code"`
		Detail []struct {
			Apellido           string   `json:"Apellido"`
			Calle              string   `json:"Calle"`
			CantidadPaquetes   string   `json:"CantidadPaquetes"`
			CodigoPostal       string   `json:"CodigoPostal"`
			CodigoPostalRetiro string   `json:"CodigoPostalRetiro"`
			Depto              struct{} `json:"Depto"`
			DeptoRetiro        struct{} `json:"DeptoRetiro"`
			DomicilioRetiro    string   `json:"DomicilioRetiro"`
			IDPieza            string   `json:"IdPieza"`
			Localidad          string   `json:"Localidad"`
			LocalidadRetiro    string   `json:"LocalidadRetiro"`
			Nombre             string   `json:"Nombre"`
			Numero             string   `json:"Numero"`
			NumeroEnvio        string   `json:"NumeroEnvio"`
			NumeroRetiro       string   `json:"NumeroRetiro"`
			PciaRetiro         string   `json:"PciaRetiro"`
			Piso               struct{} `json:"Piso"`
			PisoRetiro         struct{} `json:"PisoRetiro"`
			Provincia          string   `json:"Provincia"`
			Remito             string   `json:"Remito"`
		} `json:"detail"`
		Log []DetailLog `json:"log"`
	} `json:"data"`
	Success bool `json:"success"`
}

func createDatabase(cacheFilename string) error {
	var err error
	if cacheFilename == "" {
		cacheFilename, err = homedir.Dir()
		if err != nil {
			panic(err)
		}
		cacheFilename += "/.gocafier.db"
	}

	log.LogStd(fmt.Sprintf("Setting cache file to %s ", cacheFilename), true)

	config := &bolt.Options{Timeout: 1 * time.Second}
	appdb, err = bolt.Open(cacheFilename, 0600, config)
	if err != nil {
		panic(err)
	}
	open = true
	return nil
}

func closeDatabase() {
	open = false
	appdb.Close()
}

//Close gracefully closes the database
func Close() {
	closeDatabase()
}

//Save records a package details to the caching database
func (p *OcaPackageDetail) Save() error {
	err := appdb.Update(func(tx *bolt.Tx) error {
		packages := tx.Bucket([]byte(bucketName))

		enc, err := p.encode()
		if err != nil {
			return fmt.Errorf("could not encode response %s: %s", p.Data[0].Code, err)
		}

		err = packages.Put([]byte(p.Data[0].Code), enc)
		return err
	})
	return err
}

func (p *OcaPackageDetail) encode() ([]byte, error) {
	enc, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	return enc, nil
}

func decode(data []byte) (*OcaPackageDetail, error) {
	var p *OcaPackageDetail
	err := json.Unmarshal(data, &p)
	if err != nil {
		return nil, err
	}
	return p, nil
}

//ListPackages gets a list of packages from the database
func ListPackages() {
	appdb.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(bucketName)).Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			log.LogStd(fmt.Sprintf("key=%s", k), true)
			log.LogStd(fmt.Sprintf("value=%s\n", pretty.Formatter(v)), true)
		}
		return nil
	})
}

//GetPackage returns a single package by code
func GetPackage(code string) (*OcaPackageDetail, error) {
	if !open {
		return nil, fmt.Errorf("db must be opened before saving")
	}
	var p *OcaPackageDetail
	err := appdb.View(func(tx *bolt.Tx) error {
		var err error
		bucket := tx.Bucket([]byte(bucketName))
		key := []byte(code)
		value := bucket.Get(key)
		if value == nil {
			return nil
		}
		p, err = decode(value)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Could not get Package ID %s", code)
		return nil, err
	}
	return p, nil
}

// CreateBucket adds the application bucket to the caching database
func CreateBucket(cacheFilename string) {
	createDatabase(cacheFilename)
	appdb.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
}

//DiffWith compares the structure of the package log with another package
func (p *OcaPackageDetail) DiffWith(packageDetails OcaPackageDetail) ([]DetailLog, bool) {
	var diff []DetailLog
	foundFlag := false

	firstLog := p.Data[0].Log
	anotherLog := packageDetails.Data[0].Log

	// Loop two times, first to find slice1 strings not in slice2,
	// second loop to find slice2 strings not in slice1
	for i := 0; i < 2; i++ {
		for _, s1 := range firstLog {
			found := false
			for _, s2 := range anotherLog {
				if s1 == s2 {
					found = true
					break
				}
			}
			// String not found. We add it to return slice
			if !found {
				diff = append(diff, s1)
				foundFlag = true
			}
		}
		// Swap the slices, only if it was the first loop
		if i == 0 {
			firstLog, anotherLog = anotherLog, firstLog
		}
	}
	return diff, foundFlag
}

// SetAppDb sets the database for caching
func SetAppDb(db *bolt.DB) {
	appdb = db
}

/*func FillDatabase(listApps []App) {
	for _, app := range listApps {
		appdb.Update(func(tx *bolt.Tx) error {
			b, err := tx.CreateBucketIfNotExists([]byte(bucketName))
			if err != nil {
				return fmt.Errorf("create bucket: %s", err)
			}

			serialize, err := json.Marshal(app)

			if err != nil {
				return fmt.Errorf("Error Marshaling data: %s", err)
			}
			err = b.Put([]byte(app.Guid), serialize)

			if err != nil {
				return fmt.Errorf("Error inserting data: %s", err)
			}
			return nil
		})

	}

}

func GetAppByGuid(appGuid string) []App {
	var apps []App
	app := gcfClient.AppByGuid(appGuid)
	apps = append(apps, App{app.Name, app.Guid, app.SpaceData.Entity.Name, app.SpaceData.Entity.Guid, app.SpaceData.Entity.OrgData.Entity.Name, app.SpaceData.Entity.OrgData.Entity.Guid})
	FillDatabase(apps)
	return apps

}

func GetAllApp() []App {

	log.LogStd("Retrieving Apps for Cache...", false)
	var apps []App

	defer func() {
		if r := recover(); r != nil {
			log.LogError("Recovered in caching.GetAllApp()", r)
		}
	}()

	for _, app := range gcfClient.ListApps() {
		log.LogStd(fmt.Sprintf("App [%s] Found...", app.Name), false)
		apps = append(apps, App{app.Name, app.Guid, app.SpaceData.Entity.Name, app.SpaceData.Entity.Guid, app.SpaceData.Entity.OrgData.Entity.Name, app.SpaceData.Entity.OrgData.Entity.Guid})
	}

	FillDatabase(apps)

	log.LogStd(fmt.Sprintf("Found [%d] Apps!", len(apps)), false)

	return apps
}

func GetAppInfo(appGuid string) App {

	defer func() {
		if r := recover(); r != nil {
			log.LogError(fmt.Sprintf("Recovered from panic retrieving App Info for App Guid: %s", appGuid), r)
		}
	}()

	var d []byte
	var app App
	appdb.View(func(tx *bolt.Tx) error {
		log.LogStd(fmt.Sprintf("Looking for App %s in Cache!\n", appGuid), false)
		b := tx.Bucket([]byte(bucketName))
		d = b.Get([]byte(appGuid))
		return nil
	})
	err := json.Unmarshal([]byte(d), &app)
	if err != nil {
		return App{}
	}
	return app
}

func SetCfClient(cfClient *cfClient.Client) {
	gcfClient = cfClient

}
*/
