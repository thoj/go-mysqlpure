package main

import (
	"./mysql";
	"fmt";
	"os";
)


func main() {
	dbh, err := mysql.Connect("127.0.0.1:3306", "test", "test", "test");
	if err != nil {
		fmt.Printf("%s\n", err);
		os.Exit(1);
	}
	fmt.Printf("Connected to %s\n", dbh.ServerVersion);
	var res * mysql.MySQLResponse;
	res, err = dbh.Query("SHOW PROCESSLIST");
	res, err = dbh.Query("SELECT * FROM test");
	if err != nil {
		fmt.Printf("%s\n", err);
		os.Exit(1);
	}
	fmt.Printf("%#v\n", res);
}
