package main

import (
	"./mysql";
	"fmt";
)


func main() {
	dbh, err := mysql.Connect("127.0.0.1:3306", "test", "test", "test");
	if err != nil {
		fmt.Printf("%s\n", err);
	}
	fmt.Printf("Connected to %s\n", dbh.ServerVersion);
	err = dbh.Query("SHOW PROCESSLIST");
	if err != nil {
		fmt.Printf("%s\n", err);
	}
}
