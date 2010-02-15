package main

import (
	"mysql"
	"fmt"
	"os"
)


func main() {
	dbh, err := mysql.Connect("tcp", "", "127.0.0.1:3306", "test", "test", "") // With password, user test, with database.
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
	dbh.Use("test")
	dbh.Query("SET NAMES utf8");
	sth := new(mysql.MySQLStatement)
	sth, err = dbh.Prepare("SELECT * FROM test WHERE id = ? OR id = ? OR id = ?")
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
	res, err := sth.Execute(1, 2, 3)
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
	rows := res.FetchAllRowMap()
	fmt.Printf("Rows: %d\n", len(rows))
	fmt.Printf("%#v\n", rows)
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}

	dbh.Quit()
}
