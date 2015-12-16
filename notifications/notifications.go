package notifications

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/eljuanchosf/gocafier/caching"
	log "github.com/eljuanchosf/gocafier/logging"
	"github.com/eljuanchosf/gocafier/settings"
	"gopkg.in/gomail.v2"
)

func loadBodyTemplate(packageNumber string, packageDetail caching.OcaPackageDetail, diff []caching.DetailLog) string {
	var fullBody bytes.Buffer

	type emailData struct {
		PackageNumber string
		From          string
		Movements     []caching.DetailLog
	}

	originDetails := packageDetail.Data[0].Detail[0]

	var packageData emailData
	packageData.PackageNumber = packageNumber
	packageData.From = fmt.Sprintf("%s %s, %s, %s",
		strings.TrimSpace(originDetails.DomicilioRetiro),
		strings.TrimSpace(originDetails.NumeroRetiro),
		strings.TrimSpace(originDetails.LocalidadRetiro),
		strings.TrimSpace(originDetails.PciaRetiro))

	packageData.Movements = diff
	t, _ := template.ParseFiles("email-template.html")
	t.Execute(&fullBody, packageData)
	return fullBody.String()
}

//Send sends the email notification
func Send(currentData caching.OcaPackageDetail, diff []caching.DetailLog, smtpUser string, smtpPassword string) error {
	packageNumber := currentData.Data[0].Code

	if diff == nil {
		diff = currentData.Data[0].Log
	}

	log.LogPackage(packageNumber, "Sending notification...")
	m := gomail.NewMessage()
	m.SetHeader("From", settings.Values.Email.From)
	m.SetHeader("To", settings.Values.Email.To)
	m.SetHeader("Subject", fmt.Sprintf(settings.Values.Email.Subject, packageNumber))
	m.SetBody("text/html", loadBodyTemplate(packageNumber, currentData, diff))

	d := gomail.NewPlainDialer(settings.Values.SMTP.Server, settings.Values.SMTP.Port, smtpUser, smtpPassword)
	if err := d.DialAndSend(m); err != nil {
		return err
	}

	log.LogPackage(packageNumber, "Notification sent")
	return nil
}
