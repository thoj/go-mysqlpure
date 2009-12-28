package main

import (
	"mysql"
	"fmt"
	"os"
)


func main() {
	dbh, err := mysql.Connect("tcp", "", "127.0.0.1:3306", "predb", "predb", "predb") // With password, user test, with database.
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
	sth := new(mysql.MySQLStatement)
	//sth, err = dbh.Prepare("SELECT sha1 FROM test WHERE id = ? OR id = ? OR id = ?");
	//sth, err = dbh.Prepare("SELECT * FROM trace WHERE releaseid = ? ORDER BY date LIMIT 8");
	sth, err = dbh.Prepare("SELECT * FROM trace ORDER BY date LIMIT 10")
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
	res, err := sth.Execute()
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
	row := res.FetchAllRowMap()
	fmt.Printf("%s\n", row)
	fmt.Printf("%d\n", len(row))
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}

	dbh.Quit()
}
