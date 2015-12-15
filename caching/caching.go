package caching

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/boltdb/bolt"
)

const (
	bucketName = "packages"
)

var appdb *bolt.DB
var open bool

// OcaPackageDetail represents the response from the OCA web service
type OcaPackageDetail struct {
	Data []struct {
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
		Log []struct {
			Date        string `json:"date"`
			Description string `json:"description"`
		} `json:"log"`
	} `json:"data"`
	Success bool `json:"success"`
}

func Open(cacheFilename string) error {
	var err error
	config := &bolt.Options{Timeout: 1 * time.Second}
	appdb, err = bolt.Open(dbfile, 0600, config)
	if err != nil {
		log.Fatal(err)
	}
	open = true
	return nil
}

func Close() {
	open = false
	appdb.Close()
}

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

//GetPackage retrieves a package from the database
func GetPackage(id string) (*OcaPackageDetail, error) {
	var p *OcaPackageDetail
	err := appdb.View(func(tx *bolt.Tx) error {
		var err error
		b := tx.Bucket([]byte(bucketName))
		k := []byte(id)
		p, err = decode(b.Get(k))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Could not get Package ID %s", id)
		return nil, err
	}
	return p, nil
}

func ListPackages() {
	appdb.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(bucketName)).Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			fmt.Printf("key=%s, value=%s\n", k, v)
		}
		return nil
	})
}

// CreateBucket adds the application bucket to the caching database
func CreateBucket() {
	appdb.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil

	})
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
