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
	_, err = dbh.Query("SHOW PROCESSLIST");
	row := dbh.GetRow();
	for i := 0; i < len(row.Data); i++ {	
		fmt.Printf("%s\t", dbh.CurrentResultSet.Fields[i]);
	}
	fmt.Printf("\n");
	for i := 0; i < len(row.Data); i++ {	
		fmt.Printf("%s\t", row.Data[i]);
	}
	fmt.Printf("\n");
	if err != nil {
		fmt.Printf("%s\n", err);
		os.Exit(1);
	}
}
