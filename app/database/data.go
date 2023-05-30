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
	db *sql.DB
}

func NewDatabase() (*Database, error) {
	d := &Database{
		db: Db,
	}

	return d, nil
}

func (d *Database) SaveData(data convert.AssetSummary) error {
	query := `INSERT INTO asset_summary (time, total_jpy) VALUES (?, ?)`
	res, err := d.db.Exec(query, data.Time.Format(time.RFC3339), data.Total_JPY)
	if err != nil {
		log.Println("Error executing query:", err)
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		log.Println("Error getting last insert id:", err)
		return err
	}

	for symbol, amountJPY := range data.GMOCoinAssets_JPY {
		query = `INSERT INTO gmocoin_assets (asset_summary_id, symbol, amount_jpy) VALUES (?, ?, ?)`
		_, err = d.db.Exec(query, id, symbol, amountJPY)
		if err != nil {
			log.Println("Error executing query:", err)
			return err
		}
	}

	for symbol, amountJPY := range data.BittradeAssets_JPY {
		query = `INSERT INTO bittrade_assets (asset_summary_id, symbol, amount_jpy) VALUES (?, ?, ?)`
		_, err = d.db.Exec(query, id, symbol, amountJPY)
		if err != nil {
			log.Println("Error executing query:", err)
			return err
		}
	}

	return nil
}

type AssetSummaryRow struct {
	ID       int64
	Time     time.Time
	TotalJPY int
}

func (d *Database) GetLatestData() (convert.AssetSummary, error) {
	var latestData convert.AssetSummary
	var latestSummaryRow AssetSummaryRow

	query := `SELECT * FROM asset_summary ORDER BY time DESC LIMIT 1`
	err := d.db.QueryRow(query).Scan(&latestSummaryRow.ID, &latestSummaryRow.Time, &latestSummaryRow.TotalJPY)
	if err != nil {
		if err == sql.ErrNoRows {
			return convert.AssetSummary{}, nil
		}
		log.Println("Error executing query:", err)
		return convert.AssetSummary{}, err
	}

	latestData.Time = latestSummaryRow.Time
	latestData.Total_JPY = latestSummaryRow.TotalJPY

	query = `SELECT symbol, amount_jpy FROM gmocoin_assets WHERE asset_summary_id = ?`
	rows, err := d.db.Query(query, latestSummaryRow.ID)
	if err != nil {
		log.Println("Error executing query:", err)
		return convert.AssetSummary{}, err
	}
	defer rows.Close()

	latestData.GMOCoinAssets_JPY = make(map[string]int)
	for rows.Next() {
		var symbol string
		var amountJPY int
		err = rows.Scan(&symbol, &amountJPY)
		if err != nil {
			log.Println("Error scanning row:", err)
			return convert.AssetSummary{}, err
		}
		latestData.GMOCoinAssets_JPY[symbol] = amountJPY
	}

	query = `SELECT symbol, amount_jpy FROM bittrade_assets WHERE asset_summary_id = ?`
	rows, err = d.db.Query(query, latestSummaryRow.ID)
	if err != nil {
		log.Println("Error executing query:", err)
		return convert.AssetSummary{}, err
	}
	defer rows.Close()

	latestData.BittradeAssets_JPY = make(map[string]int)
	for rows.Next() {
		var symbol string
		var amountJPY int
		err = rows.Scan(&symbol, &amountJPY)
		if err != nil {
			log.Println("Error scanning row:", err)
			return convert.AssetSummary{}, err
		}
		latestData.BittradeAssets_JPY[symbol] = amountJPY
	}

	return latestData, nil
}
