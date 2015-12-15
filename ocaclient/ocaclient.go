package ocaclient

import (
	"encoding/json"

	"github.com/ddliu/go-httpclient"
	"github.com/eljuanchosf/gocafier/caching"
)

const (
	ocaBaseURL = "http://www.oca.com.ar"
)

// RequestData sends a GET request to the OCA web service using
// the packageType and packageNumber provided by the user.
func RequestData(packageType string, packageNumber string) (response caching.OcaPackageDetail, err error) {
	httpclient.Defaults(httpclient.Map{
		httpclient.OPT_USERAGENT: "Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2228.0 Safari/537.36",
		"Accept":                 "application/json, text/javascript, */*; q=0.01",
		"Accept-Encoding":        "deflate",
		"Accept-Language":        "en-US,en;q=0.8,es;q=0.6",
		"Connection":             "keep-alive",
		"Host":                   "www.oca.com.ar",
		"Referer":                "http://www.oca.com.ar/",
		"X-Requested-With":       "XMLHttpRequest",
	})

	res, err := httpclient.Get(ocaBaseURL, map[string]string{
		"q":      "package-locator",
		"type":   packageType,
		"number": packageNumber,
	})

	defer res.Body.Close()
	decoder := json.NewDecoder(res.Body)
	var ocaData caching.OcaPackageDetail
	err = decoder.Decode(&ocaData)
	return ocaData, err
}
