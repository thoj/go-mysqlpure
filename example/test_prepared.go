package main

import (
	"mysql";
	"fmt";
	"os";
)


func main() {
	dbh, err := mysql.Connect("tcp", "", "127.0.0.1:3306", "test", "test", "test");	// With password, user test, with database.
	if err != nil {
		fmt.Printf("%s\n", err);
		os.Exit(1);
	}
	dbh.Use("test");	//Select database

	sth := new(mysql.MySQLStatement);
	sth, err = dbh.Prepare("SELECT sha1 FROM test WHERE id = ? OR id = ? OR id = ?");
	res, err := sth.Execute(2, "93", 8);
	if err != nil {
		fmt.Printf("%s\n", err);
		os.Exit(1);
	}
	for rowmap := res.FetchRowMap(); rowmap != nil; rowmap = res.FetchRowMap() {
		fmt.Printf("%#v\n", rowmap)
	}

	dbh.Quit();
}
