package mysql

import (
	"fmt"
	"testing"
)

func MakeDbh(t *testing.T) *MySQLInstance {
	dbh, err := Connect("tcp", "127.0.0.1:3306", "test", "test", "test")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if dbh == nil {
		t.Error("dbh is nil")
		t.FailNow()
	}
	return dbh
}

func CheckQuery(t *testing.T, dbh *MySQLInstance, q string) *MySQLResponse {
	res, err := dbh.Query(q)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	return res
}

func SelectSingleRow(t *testing.T, q string) map[string]string {
	dbh := MakeDbh(t)
	dbh.Use("test")

	res := CheckQuery(t, dbh, "SET NAMES utf8")
	res = CheckQuery(t, dbh, q)
	row := res.FetchRowMap()
	dbh.Quit()
	return row
}

func SelectSingleRowPrepared(t *testing.T, q string, p ...interface{}) map[string]string {
	dbh := MakeDbh(t)
	dbh.Use("test")

	res := CheckQuery(t, dbh, "SET NAMES utf8")
	sth, err := dbh.Prepare(q)

	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	res, err = sth.Execute(p...)
	if err != nil || res == nil {
		t.Error(err)
		t.FailNow()
	}
	row := res.FetchRowMap()
	dbh.Quit()
	return row
}

func TestLongRun(t *testing.T) {
	dbh := MakeDbh(t)
	for i := 0; i < 10000; i++ {
		res := CheckQuery(t, dbh, fmt.Sprintf("INSERT INTO test2 (test, testtest) VALUES(%d, %d)", i, i%10))
		if res.InsertId < 1 {
			t.Error("InsertId < 0")
			t.FailNow()
		}
		if i%1000 == 0 {
			fmt.Printf("%d%%\n", i/100)
		}
	}
	res := CheckQuery(t, dbh, "DELETE FROM test2")
	if res.AffectedRows != 10000 {
		t.Error("AffectedRows = ", res.AffectedRows)
		t.FailNow()
	}
	res = CheckQuery(t, dbh, "TRUNCATE TABLE test2")
}

func TestUnfinished(t *testing.T) {
	dbh := MakeDbh(t)
	res := CheckQuery(t, dbh, "SELECT * FROM test")
	row := res.FetchRowMap()
	res = CheckQuery(t, dbh, "SELECT * FROM test WHERE name='test1'")
	row = res.FetchRowMap()
	test := "1234567890abcdef"
	if row == nil || row["stuff"] != test {
		t.Error(row["stuff"], " != ", test)
	}
	dbh.Quit()
}

func TestSelectString(t *testing.T) {
	row := SelectSingleRow(t, "SELECT * FROM test WHERE name='test1'")
	test := "1234567890abcdef"
	if row == nil || row["stuff"] != test {
		t.Error(row["stuff"], " != ", test)
	}
}

func TestSelectStringPrepared2(t *testing.T) {
	row := SelectSingleRowPrepared(t, "SELECT * FROM test WHERE name=? OR name=?", "test1", "testnonexisting")
	test := "1234567890abcdef"
	if row == nil || row["stuff"] != test {
		t.Error(row["stuff"], " != ", test)
	}
}

func TestSelectStringPrepared(t *testing.T) {
	row := SelectSingleRowPrepared(t, "SELECT * FROM test WHERE name=?", "test1")
	test := "1234567890abcdef"
	if row == nil || row["stuff"] != test {
		t.Error(row["stuff"], " != ", test)
	}
}

func TestSelectUFT8(t *testing.T) {
	row := SelectSingleRow(t, "SELECT * FROM test WHERE name='unicodetest1'")
	test := "l̡̡̡ ̴̡ı̴̴̡ ̡̡͡|̲̲̲͡͡͡ ̲▫̲͡ ̲̲̲͡͡π̲̲͡͡ ̲̲͡▫̲̲͡͡ ̲|̡̡̡ ̡ ̴̡ı̴̡̡ ̡"
	if row == nil || row["stuff"] != test {
		t.Error(row["stuff"], " != ", test)
	}
}

func TestSelectUFT8Prepared(t *testing.T) {
	row := SelectSingleRowPrepared(t, "SELECT * FROM test WHERE name=?", "unicodetest1")
	test := "l̡̡̡ ̴̡ı̴̴̡ ̡̡͡|̲̲̲͡͡͡ ̲▫̲͡ ̲̲̲͡͡π̲̲͡͡ ̲̲͡▫̲̲͡͡ ̲|̡̡̡ ̡ ̴̡ı̴̡̡ ̡"
	if row == nil || row["stuff"] != test {
		t.Error(row["stuff"], " != ", test)
	}
}

func TestSelectEmpty(t *testing.T) {
	row := SelectSingleRowPrepared(t, "SELECT * FROM test WHERE name='doesnotexist'")
	if row != nil {
		t.Error("Row is not nil")
	}
}

func TestError(t *testing.T) {
	dbh := MakeDbh(t)
	dbh.Use("test")

	res, err := dbh.Query("SELECT * FROM test WHERE namefail='foo'")
	if res != nil || err == nil {
		t.Error("err == nil, expected error")
	}
	dbh.Quit()
}
