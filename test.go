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
	res := new(mysql.MySQLResponse);
//	res, err = dbh.Query("SELECT * FROM releases LIMIT 1000");
	res, err = dbh.Query("INSERT INTO test (sha1) VALUES(SHA1('foo'))");
	fmt.Printf("%s\n", res);
	if err != nil {
		fmt.Printf("%s\n", err);
		os.Exit(1);
	}
	res, err = dbh.Query(fmt.Sprintf("DELETE FROM test WHERE id=%d", res.InsertId));
	fmt.Printf("%s\n", res);
/*
	for row := dbh.FetchRow(); row != nil; row = dbh.FetchRow() {
		for i := 0; i < len(row.Data); i++ {
			fmt.Printf("%s\t", dbh.CurrentResultSet.Fields[i])
		}
		fmt.Printf("\n");
		for i := 0; i < len(row.Data); i++ {
			fmt.Printf("%s\t", row.Data[i])
		}
		fmt.Printf("\n");
		if err != nil {
			fmt.Printf("%s\n", err);
			os.Exit(1);
		}
	}
*/
	dbh.Quit();
}
