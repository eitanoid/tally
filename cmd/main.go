package main

import (
	"database/sql"
	"fmt"

	_ "github.com/ncruces/go-sqlite3/driver"
)

func main() {
	var version string
	db, _ := sql.Open("sqlite3", "file:my.db")
	defer fmt.Println(db.Close())
	db.QueryRow(`SELECT sqlite_version()`).Scan(&version)
	fmt.Println(version)
}
