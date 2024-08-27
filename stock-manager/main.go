package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
	"github.com/nats-io/nats.go"
)

// Message represents the structure of the data received from NATS
type Message struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

func main() {
	// Connect to SQLite database
	db, err := sql.Open("sqlite3", "./data.db")
	if err != nil {
		log.Fatal("Failed to open SQLite database:", err)
	}
	defer db.Close()

	// Create a table if it doesn't exist
	createTableSQL := `CREATE TABLE IF NOT EXISTS data (
		"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		"name" TEXT,
		"value" TEXT
	);`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatal("Failed to create table:", err)
	}

	// Connect to NATS server
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal("Failed to connect to NATS:", err)
	}
	defer nc.Close()

	// Subscribe to a NATS subject
	sub, err := nc.Subscribe("data.subject", func(m *nats.Msg) {
		log.Printf("Received a message: %s", string(m.Data))

		// Process the received message and store it in the database
		err := processAndStore(db, m.Data)
		if err != nil {
			log.Println("Error processing and storing data:", err)
		}
	})
	if err != nil {
		log.Fatal("Failed to subscribe to subject:", err)
	}

	// Keep the service running
	log.Println("Service is running... waiting for messages.")
	select {} // Block forever

	// Unsubscribe when finished
	defer sub.Unsubscribe()
}

// Process and store the received data in the SQLite database
func processAndStore(db *sql.DB, data []byte) error {
	// For simplicity, assume the data is a simple comma-separated string: "name,value"
	var name, value string
	n, err := fmt.Sscanf(string(data), "%s,%s", &name, &value)
	if err != nil || n != 2 {
		return fmt.Errorf("invalid data format")
	}

	// Insert data into the SQLite database
	insertSQL := `INSERT INTO data (name, value) VALUES (?, ?)`
	_, err = db.Exec(insertSQL, name, value)
	if err != nil {
		return fmt.Errorf("failed to insert data into database: %v", err)
	}

	log.Println("Data successfully stored in the database")
	return nil
}
