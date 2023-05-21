package database

import (
	"assetmanagement/appconfig"
	"assetmanagement/convert"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var Db *sql.DB  


func init() {
	cmd := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=UTC",
		appconfig.AppConfig.Dbuser,
		appconfig.AppConfig.Dbpass,
		appconfig.AppConfig.Dblocalhost,
		appconfig.AppConfig.Dbport,
		appconfig.AppConfig.Dbname,
	)
	var err error
	Db, err = sql.Open("mysql", cmd)  
	if err != nil {
		log.Fatal(err)  
	}
}


type Database struct {
	db	*sql.DB
}

func NewDatabase() (*Database, error) {
	d := &Database{
		db:	Db,
	}

	return d, nil
}

func (d *Database) SaveData() error {
	data, err := convert.FetchDataSummary()
	if err != nil {
		return err
	}

	var lastData convert.AssetSummary
	var timestamp string
	err = d.db.QueryRow("SELECT * FROM asset_summary ORDER BY time DESC LIMIT 1").Scan(
		&timestamp,
		&lastData.BTC_JPY,
		&lastData.XEM_JPY,
		&lastData.ADA_JPY,
		&lastData.Total_JPY,
	)
	
	if err != nil && err != sql.ErrNoRows {
		log.Println("Error getting last data:", err)
		return err
	}


	lastData.Time, err = time.Parse(time.RFC3339, timestamp) 
	if err != nil {
		log.Println("Error parsing time:", err)
	}

	if data.BTC_JPY != lastData.BTC_JPY || data.XEM_JPY != lastData.XEM_JPY || data.ADA_JPY != lastData.ADA_JPY {
		query := `INSERT INTO asset_summary (time, btc_jpy, xem_jpy, ada_jpy, total_jpy) VALUES (?, ?, ?, ?, ?)`
		_, err = d.db.Exec(query, data.Time.Format(time.RFC3339), data.BTC_JPY, data.XEM_JPY, data.ADA_JPY, data.Total_JPY)
		if err != nil {
			log.Println("Error executing query:", err)
			return err
		}
	}

	return nil
}