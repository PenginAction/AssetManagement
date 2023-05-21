package main

import (
	"assetmanagement/app/webserver"
	"assetmanagement/appconfig"
	"assetmanagement/utils"
)

func main() {
	utils.SetupLogging(appconfig.AppConfig.LogFile)
	webserver.Start()
}
