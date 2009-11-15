package main

import (
	"./mysql";
	"fmt";
)


func main() {
	err := mysql.Connect("127.0.0.1:3306", "test", "test", "test");
	if err != nil {
		fmt.Printf("%s\n", err);
	}
}
