package main

import (
	"authentication/data"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

const webPort = "80"

var counts int64

type Config struct {
	DB     *sql.DB
	Models data.Models
}

func main() {
	log.Println("Dtarting authentication service")

	//connect to DB
	conn := connectToDB()
	if conn == nil {
		log.Panic("Cant connect to Postgres!")
	}

	//set up config
	app := Config{
		DB:     conn,
		Models: data.New(conn),
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	err := srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)

	if err != nil {
		return nil, err
	}

	err = db.Ping()

	if err != nil {
		return nil, err
	}

	var currentDB string
	err = db.QueryRow("SELECT current_database()").Scan(&currentDB)
	if err != nil {
		log.Fatal("Failed to check current database:", err)
	}
	log.Println("✅ Connected to database:", currentDB)

	var ip, dbname string
	err = db.QueryRow("SELECT inet_server_addr(), current_database()").Scan(&ip, &dbname)
	if err != nil {
		log.Fatal("❌ Failed to get connection info:", err)
	}
	log.Printf("✅ Go is connected to DB '%s' on IP: %s\n", dbname, ip)

	return db, nil
}

func connectToDB() *sql.DB {
	dsn := os.Getenv("DSN")
	log.Println("dsn is ", dsn)
	for {
		connection, err := openDB(dsn)
		if err != nil {
			log.Println("Postgres not yet ready ...")
			counts++
		} else {
			log.Println("Connected to Postgres!")
			return connection
		}

		if counts > 10 {
			log.Println(err)
			return nil
		}
		log.Println("Backing off for two seconds ...")
		time.Sleep(2 * time.Second)
		continue

	}
}
