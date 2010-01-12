package mysql

import (
	"testing"
//	"testing/quick"
)

func TestSelectString(t *testing.T) {
	dbh, err := Connect("tcp", "", "127.0.0.1:3306", "test", "", "test")
	if err != nil {
		t.Error(err)
	}
	if dbh == nil {
		t.Error("dbh is nil")
	}
	dbh.Use("test")

	res, err := dbh.Query("SELECT * FROM test WHERE name='test1'")

	if err != nil {
		t.Error(err)
	}
	row := res.FetchRowMap()
	if row == nil {
		t.Error("row is nil")
	}
	dbh.Quit()
	if row["stuff"] == "1234567890abcdef" {
		t.Error(row["stuff"])
	}
}
