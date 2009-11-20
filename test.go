package main

import (
	"./mysql";
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

	res := new(mysql.MySQLResponse);
	res, err = dbh.Query("INSERT INTO test (sha1) VALUES(SHA1('foo'))");

	fmt.Printf("%s\n", res);
	if err != nil {
		fmt.Printf("%s\n", err);
		os.Exit(1);
	}
	res, err = dbh.Query(fmt.Sprintf("DELETE FROM test WHERE id=%d", res.InsertId));
	fmt.Printf("%s\n", res);
	res, err = dbh.Query("SELECT * FROM test");
	fmt.Printf("-----%s\n", res);
	for rowmap := res.FetchRowMap(); rowmap != nil; rowmap = res.FetchRowMap() {
		fmt.Printf("%#v\n", rowmap);
	}
	res, err = dbh.Query("SHOW PROCESSLIST");
	fmt.Printf("%s %s\n", res, err);
	for row := res.FetchRow(); row != nil; row = res.FetchRow() {
		for i := 0; i < len(row.Data); i++ {
			fmt.Printf("%s\t", res.ResultSet.Fields[i])
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
	dbh.Quit();
}
