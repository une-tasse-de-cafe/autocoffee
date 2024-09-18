package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

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

	db, err := sql.Open("sqlite3", "./data.db")
	if err != nil {
		log.Fatal("Failed to open SQLite database:", err)
	}
	defer db.Close()

	// Create a table if it doesn't exist
	createTableSQL := `CREATE TABLE IF NOT EXISTS data (
		type TEXT NOT NULL PRIMARY KEY,
		value TEXT
	);`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatal("Failed to create table:", err)
	}

	// Check and insert missing rows if they don't exist
	rows := []struct {
		Type  string
		Value string
	}{
		{"mixed", "150"},
		{"arabica", "150"},
		{"robusta", "150"},
	}

	for _, row := range rows {
		var exists bool
		query := `SELECT EXISTS(SELECT 1 FROM data WHERE type=? LIMIT 1)`
		err = db.QueryRow(query, row.Type).Scan(&exists)
		if err != nil {
			log.Fatal("Failed to query database:", err)
		}

		if !exists {
			insertSQL := `INSERT INTO data (type, value) VALUES (?, ?)`
			_, err = db.Exec(insertSQL, row.Type, row.Value)
			if err != nil {
				log.Fatal("Failed to insert row:", err)
			}
			log.Printf("📫 Inserted missing row: type=%s, value=%s\n", row.Type, row.Value)
		} else {
			log.Printf("📫 Row already exists: type=%s\n", row.Type)
		}
	}

	opt, err := nats.NkeyOptionFromSeed("seed.txt")
	if err != nil {
		log.Fatal(err)
	}

	nc, err := nats.Connect(os.Getenv("NATS_URL"), opt)
	if err != nil {
		log.Fatal("connect to nats: ", err)
	}

	defer nc.Close()

	sub, err := nc.Subscribe("coffee.stock.>", func(m *nats.Msg) {
		log.Printf("📥 Received a message: %s \n", string(m.Data))
		subjectUri := strings.Split(m.Subject, ".")

		if len(subjectUri) != 4 && len(subjectUri) != 5 {
			log.Println("The subject length isn't correct")
			return
		}

		// Authorize only requests for knowned type/action
		typeBean := subjectUri[2]
		switch typeBean {
		case "arabica", "robusta", "mixed":
			break
		default:
			log.Println("Not supported")
			return
		}

		action := subjectUri[3]
		switch action {
		// All message are answered with the number of coffee left
		// coffee.stock.robusta.get
		case "get":
			getRequest := `SELECT value FROM data WHERE type=?`
			var value string
			err = db.QueryRow(getRequest, typeBean).Scan(&value)
			if err != nil {
				log.Println("💢 Cannot requests to the db")
				return
			}
			log.Println("Coffee left: " + value)
			err = m.Respond([]byte(value))
			if err != nil {
				log.Println("💢 Can't respond to client : " + err.Error())
				return
			}

		// coffee.stock.robusta.dec.large
		case "dec":
			if len(subjectUri) != 5 {
				log.Println("💢 Bad request")
				return
			}
			size := subjectUri[4]

			coffeeQuantityMap := make(map[string]string)
			coffeeQuantityMap["small"] = "9"
			coffeeQuantityMap["medium"] = "17"
			coffeeQuantityMap["large"] = "30"

			quantity, ok := coffeeQuantityMap[size]
			if !ok {
				fmt.Println("Invalid coffee size")
				return
			}

			updateQuery := `UPDATE data SET value = value - ? WHERE type = ?`
			_, err = db.Exec(updateQuery, quantity, typeBean)
			if err != nil {
				fmt.Println("💢 Failed to update coffee quantity")
				return
			}
			log.Printf("Decremented %s coffee by %s", typeBean, quantity)
			m.Respond([]byte("OK"))

		default:
			fmt.Println("💢 This action is not supported.")
			return
		}

		// Display the remaining quantity of all coffee types
		rows, err := db.Query("SELECT type, value FROM data")
		if err != nil {
			fmt.Println("💢 Failed to fetch coffee data")
			return
		}
		defer rows.Close()

		log.Println("📦 Remaining coffee quantities:")
		for rows.Next() {
			var coffeeType, quantity string
			err := rows.Scan(&coffeeType, &quantity)
			if err != nil {
				fmt.Println("💢 Failed to scan coffee data")
				return
			}
			fmt.Printf("☕ %s: %s\n", coffeeType, quantity)
		}

	})

	if err != nil {
		log.Fatal("💢 Failed to subscribe to subject:", err)
	}

	// Keep the service running
	log.Println("⌛ Waiting for orders...")

	defer sub.Unsubscribe()
	select {} // Block forever

}
