package logging

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
)

var (
	debugFlag bool
)

func Connect() bool {

	success := false
	logrus.SetFormatter(&logrus.JSONFormatter{})
	if !debugFlag {
		logrus.SetOutput(ioutil.Discard)
	} else {
		logrus.SetOutput(os.Stdout)
	}
	return success
}

func SetupLogging(debug bool) {
	debugFlag = debug
}

func LogPackage(packageNumber string, message string) {
	LogStd(fmt.Sprintf("P:%s - %s", packageNumber, message), true)
}

func LogStd(message string, force bool) {
	Log(message, force, false, nil)
}

func LogError(message string, errMsg interface{}) {
	Log(message, false, true, errMsg)
}

func Log(message string, force bool, isError bool, err interface{}) {

	if debugFlag || force || isError {

		writer := os.Stdout
		var formattedMessage string

		if isError {
			writer = os.Stderr
			formattedMessage = fmt.Sprintf("[%s] Exception occurred! Message: %s Details: %v", time.Now().String(), message, err)
		} else {
			formattedMessage = fmt.Sprintf("[%s] %s", time.Now().String(), message)
		}
		fmt.Fprintln(writer, formattedMessage)
	}
}
