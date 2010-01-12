package mysql

import (
	"testing"
//	"testing/quick"
)

func SelectSingleRow(t *testing.T, q string) map[string]string {
	dbh, err := Connect("tcp", "", "127.0.0.1:3306", "test", "", "test")
	if err != nil {
		t.Error(err)
	}
	if dbh == nil {
		t.Error("dbh is nil")
	}
	dbh.Use("test")
	
	res, err := dbh.Query("SET NAMES utf8")
	res, err = dbh.Query(q)

	if err != nil {
		t.Error(err)
	}
	row := res.FetchRowMap()
	if row == nil {
		t.Error("row is nil")
	}
	dbh.Quit()
	return row;
}

func TestSelectString(t *testing.T) {
	row := SelectSingleRow(t, "SELECT * FROM test WHERE name='test1'");
	if row["stuff"] != "1234567890abcdef" {
		t.Error(row["stuff"])
	}
}

func TestSelectUFT8(t *testing.T) {
	row := SelectSingleRow(t, "SELECT * FROM test WHERE name='unicodetest1'");
	if row["stuff"] != "l̡̡̡ ̴̡ı̴̴̡ ̡̡͡|̲̲̲͡͡͡ ̲▫̲͡ ̲̲̲͡͡π̲̲͡͡ ̲̲͡▫̲̲͡͡ ̲|̡̡̡ ̡ ̴̡ı̴̡̡ ̡" {
		t.Error(row["stuff"])
	}
}
