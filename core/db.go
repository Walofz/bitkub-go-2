package core

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB(dbPath string) error {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creating database directory: %w", err)
	}

	var err error
	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("error opening database: %w", err)
	}

	// Set database connection pool settings
	sqlcmd := `CREATE TABLE IF NOT EXISTS trades (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME,
		asset TEXT,
		operation TEXT, 
		amount_thb REAL,
		coin_amount REAL,
		price REAL,
		mode TEXT,
		deviation REAL,
		log_message TEXT)`

	_, err = DB.Exec(sqlcmd)
	if err != nil {
		return fmt.Errorf("error creating trades table: %w", err)
	}

	fmt.Println("✅ Database initialized at:", dbPath)
	return nil
}

func LogTrade(asset string, operation string, amountTHB float64, coinAmount float64, price float64, mode string, deviation float64, logMessage string) {
	if DB == nil {
		fmt.Println("❌ Error: Database connection is nil. Cannot log trade.")
		return
	}

	sqlcmd := `INSERT INTO trades (timestamp, asset, operation, amount_thb, coin_amount, price, mode, deviation, log_message) 
			   VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := DB.Exec(sqlcmd, time.Now(), asset, operation, amountTHB, coinAmount, price, mode, deviation, logMessage)

	if err != nil {
		fmt.Printf("❌ Error saving trade to DB: %v\n", err)
	}
}

func GetProductionTrades(limit int) ([]TradeRecord, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	query := `
		SELECT id, timestamp, asset, operation, amount_thb, coin_amount, price, deviation
		FROM trades
		WHERE mode = 'PRODUCTION'
		ORDER BY id DESC
		LIMIT ?
	`

	rows, err := DB.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trades []TradeRecord
	for rows.Next() {
		var r TradeRecord
		var ts time.Time
	
		err := rows.Scan(&r.ID, &ts, &r.Asset, &r.Operation, &r.AmountTHB, &r.CoinAmount, &r.Price, &r.Deviation)
		if err != nil {
			continue
		}

		r.Timestamp = ts.Format("02/01/2006 15:04:05")
		trades = append(trades, r)
	}

	return trades, nil
}
