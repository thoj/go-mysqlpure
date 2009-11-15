package main

import (
	"./mysql";
)


func main() {
	_ = mysql.Connect("127.0.0.1:3306", "pre", "test", "test");
}
