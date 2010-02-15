package main

import (
	"mysql"
	"fmt"
	"os"
)


func main() {
	//dbh, err := mysql.Connect("unix", "", "/var/run/mysqld/mysqld.sock", "test", "", ""); // Without password, user test, without database.
	dbh, err := mysql.Connect("tcp", "", "127.0.0.1:3306", "test", "test", "") // With password, user test, with database.
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
	dbh.Use("test") //Select database

	res := new(mysql.MySQLResponse)

	res, err = dbh.Query("SET NAMES utf8")
	res, err = dbh.Query("INSERT INTO test (name, stuff) VALUES('testrow', '12345')")

	fmt.Printf("%s\n", res)
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
	res, err = dbh.Query(fmt.Sprintf("DELETE FROM test WHERE id=%d", res.InsertId))
	fmt.Printf("%s\n", res)
	res, err = dbh.Query("SELECT * FROM test")
	fmt.Printf("%s\n", res)

	for rowmap := res.FetchRowMap(); rowmap != nil; rowmap = res.FetchRowMap() {
		fmt.Printf("%#v\n", rowmap)
	}

	res, err = dbh.Query("SHOW PROCESSLIST")
	for rowmap := res.FetchRowMap(); rowmap != nil; rowmap = res.FetchRowMap() {
		fmt.Printf("%#v\n", rowmap)
	}

	dbh.Quit()
}
