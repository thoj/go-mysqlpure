package main

import (
	"mysql";
	"fmt";
	"os";
)


func main() {
	//dbh, err := mysql.Connect("unix", "", "/var/run/mysqld/mysqld.sock", "test", "", ""); // Without password, user test, without database.
	dbh, err := mysql.Connect("tcp", "", "127.0.0.1:3306", "test", "test", "test");	// With password, user test, with database.
	if err != nil {
		fmt.Printf("%s\n", err);
		os.Exit(1);
	}
	dbh.Use("test");	//Select database

	sth := new(mysql.MySQLStatement);
//	sth, err = dbh.Prepare("INSERT INTO test (sha1) VALUES(SHA1(?))");
	sth, err = dbh.Prepare("SELECT * FROM test WHERE id = ? OR id = ? OR id = ?");

	fmt.Printf("%s\n", sth);
	if err != nil {
		fmt.Printf("%s\n", err);
		os.Exit(1);
	}
	dbh.Quit();
}
