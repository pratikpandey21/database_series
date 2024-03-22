package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const (
	dsn         = "root:rootpassword@tcp(mysql:3306)/isolation_demo?parseTime=true"
	updateQuery = "UPDATE test_data SET value = ? WHERE id = 1"
	selectQuery = "SELECT value FROM test_data WHERE id = 1"
)

func main() {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	validateIsolationLevels(db)
}

func validateIsolationLevels(db *sql.DB) {
	// Run Transaction A in a goroutine
	go runTransactionA(db)

	// Give Transaction A a head start
	time.Sleep(1 * time.Second)

	// Run Transaction B with READ UNCOMMITTED
	runTransactionB(db, "READ UNCOMMITTED")

	revertState(db)

	go runTransactionA(db)

	// Give Transaction A a head start
	time.Sleep(1 * time.Second)
	// Run Transaction B with READ COMMITTED
	runTransactionB(db, "READ COMMITTED")

	revertState(db)

	go runTransactionA(db)

	// Give Transaction A a head start
	time.Sleep(1 * time.Second)
	// Run Transaction B with READ COMMITTED
	runTransactionB(db, "REPEATABLE READ")

	revertState(db)

	go runTransactionA(db)

	// Give Transaction A a head start
	time.Sleep(1 * time.Second)
	// Run Transaction B with READ COMMITTED
	runTransactionB(db, "SERIALIZABLE")
}

func revertState(db *sql.DB) {
	// Start Transaction A with default isolation level
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	// Update record
	_, err = tx.Exec(updateQuery, "Test")
	if err != nil {
		log.Fatal(err)
	}

	// Commit Transaction A
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Transaction reverted")
}

func runTransactionA(db *sql.DB) {
	// Start Transaction A with default isolation level
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	// Update record
	_, err = tx.Exec(updateQuery, "Updated by Transaction A")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Transaction A updated the record, sleeping for 5 seconds...")
	time.Sleep(5 * time.Second)

	// Commit Transaction A
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Transaction A committed")
}

func runTransactionB(db *sql.DB, isolationLevel string) {
	// Set the transaction isolation level for Transaction B
	if _, err := db.Exec(fmt.Sprintf("SET SESSION TRANSACTION ISOLATION LEVEL %s", isolationLevel)); err != nil {
		log.Fatal(err)
	}

	// Start Transaction B
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	// Attempt to read the updated record before Transaction A commits
	var value string
	err = tx.QueryRow(selectQuery).Scan(&value)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Transaction B read under %s before A commits: %s\n", isolationLevel, value)

	// Wait to ensure Transaction A commits
	time.Sleep(6 * time.Second)

	// Attempt to read again after Transaction A commits
	err = tx.QueryRow(selectQuery).Scan(&value)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Transaction B read under %s after A commits: %s\n", isolationLevel, value)

	_ = tx.Rollback()
}
