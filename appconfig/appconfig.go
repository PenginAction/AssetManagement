package appconfig

import (
	"log"
	"os"

	"github.com/go-ini/ini"
)

type Configuration struct {
	GmoapiKey    string
	GmoapiSecret string
	Symbol1      string
	Symbol2      string
	BitapiKey    string
	BitapiSecret string
	Symbol3      string
	LogFile      string
	Dbuser       string
	Dbpass       string
	Dblocalhost  string
	Dbport       int
	Dbname       string
	Port         int
}

var AppConfig Configuration

func init() {
	cfg, err := ini.Load("config.ini")
	if err != nil {
		log.Printf("Failed to read file: %v", err)
		os.Exit(1)
	}

	AppConfig = Configuration{
		GmoapiKey:    cfg.Section("GMOCoin").Key("gmoapi_key").String(),
		GmoapiSecret: cfg.Section("GMOCoin").Key("gmoapi_secret").String(),
		Symbol1:      cfg.Section("GMOCoin").Key("gmosymbol1").String(),
		Symbol2:      cfg.Section("GMOCoin").Key("gmosymbol2").String(),
		Symbol3:      cfg.Section("GMOCoin").Key("gmosymbol3").String(),
		BitapiKey:    cfg.Section("bittrade").Key("bitapi_key").String(),
		BitapiSecret: cfg.Section("bittrade").Key("bitapi_secret").String(),
		LogFile:      cfg.Section("assetmanagement").Key("log_file").String(),
		Dbuser:       cfg.Section("database").Key("dbuser").String(),
		Dbpass:       cfg.Section("database").Key("dbpass").String(),
		Dblocalhost:  cfg.Section("database").Key("dblocalhost").String(),
		Dbport:       cfg.Section("database").Key("dbport").MustInt(),
		Dbname:       cfg.Section("database").Key("dbname").String(),
		Port:         cfg.Section("web").Key("port").MustInt(),
	}
}
