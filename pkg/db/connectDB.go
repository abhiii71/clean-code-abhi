package db

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/abhiii71/clean-code-abhi/pkg/config"
	_ "github.com/lib/pq"
)

func ConnectPGDB(cnf config.Config) *sql.DB {
	fmt.Println(cnf)
	postgresURL := fmt.Sprintf("%v://%v:%v@%v:%v/%v?sslmode=%v", cnf.PGDBmsName, cnf.PgUserName, cnf.PgPassword, cnf.PGHost, cnf.PgPort, cnf.PGDBName, cnf.PgSSLMode)
	db, err := sql.Open(cnf.PgDriverName, postgresURL)
	if err != nil {
		log.Fatal(err, err.Error(), "driver name", cnf.PgDriverName, "postgres URL", postgresURL)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("not connected to postgres db: ", err.Error())
	}
	log.Println("Connected to postres db Successfully!")
	return db
}
