package mysql

import (
	"fmt"
	"encoding/binary"
	"bufio"
	"time"
	"os"
	"log"
)

type PacketHeader struct {
	Len uint64
	Seq uint8
}

type MySQLResultSet struct {
	FieldCount uint64
	Fields     []*MySQLField
	Rows       []*MySQLRow
}

type MySQLResponse struct {
	FieldCount   uint64
	AffectedRows uint64
	InsertId     uint64
	ServerStatus uint16
	WarningCount uint16
	Message      string
	EOF          bool
	Prepared     bool //Result from prapered statement

	ResultSet *MySQLResultSet
	mysql     *MySQLInstance
}


func (s *MySQLStatement) String() string {
	return fmt.Sprintf("Statement Id = %d, Columns = %d, Parameters = %d", s.StatementId, s.Columns, s.Parameters)
}

func (r *MySQLResponse) String() string {
	var msg string
	if r == nil {
		return "nil"
	}
	if r.FieldCount == 0x00 {
		msg = fmt.Sprintf("Response = OK")
	} else if r.FieldCount == 0xff {
		msg = fmt.Sprintf("Response = ERROR")
	} else {
		msg = fmt.Sprintf("Response = ResultSet")
	}
	if r.AffectedRows > 0 {
		msg = fmt.Sprintf("%s, Affected Rows = %d", msg, r.AffectedRows)
	}
	if r.InsertId > 0 {
		msg = fmt.Sprintf("%s, Insert Id = %d", msg, r.InsertId)
	}
	msg = fmt.Sprintf("%s, Server Status = %x", msg, r.ServerStatus)
	if r.WarningCount > 0 {
		msg = fmt.Sprintf("%s, Warnings = %x", msg, r.WarningCount)
	}
	if len(r.Message) > 0 {
		msg = fmt.Sprintf("%s, Message = %s", msg, r.Message)
	}
	return msg
}

// This is terrible should return a interface or something instead of converting to strings.
func readFieldData(br *bufio.Reader, f *MySQLField) (string, bool, os.Error) {
	switch f.Type {
	case MYSQL_TYPE_TINY:
		var l int8
		err := binary.Read(br, binary.LittleEndian, &l)
		return fmt.Sprintf("%d", l), false, err
	case MYSQL_TYPE_SHORT:
		var l int16
		err := binary.Read(br, binary.LittleEndian, &l)
		return fmt.Sprintf("%d", l), false, err
	case MYSQL_TYPE_LONG:
		var l int32
		err := binary.Read(br, binary.LittleEndian, &l)
		return fmt.Sprintf("%d", l), false, err
	case MYSQL_TYPE_LONGLONG:
		var l int64
		err := binary.Read(br, binary.LittleEndian, &l)
		return fmt.Sprintf("%d", l), false, err
	case MYSQL_TYPE_FLOAT:
		var f float32
		err := binary.Read(br, binary.LittleEndian, &f)
		return fmt.Sprintf("%f", f), false, err
	case MYSQL_TYPE_DOUBLE:
		var f float64
		err := binary.Read(br, binary.LittleEndian, &f)
		return fmt.Sprintf("%f", f), false, err
	case MYSQL_TYPE_VAR_STRING:
		return unpackString(br)
	case MYSQL_TYPE_STRING:
		return unpackString(br)
	case MYSQL_TYPE_BLOB:
		return unpackString(br)
	case MYSQL_TYPE_DATETIME:
		dt, err := unpackDateTime(br)
		return fmt.Sprintf("%s", dt), false, err
	case MYSQL_TYPE_DATE:
		dt, err := unpackDate(br)
		return fmt.Sprintf("%s", dt), false, err
	case MYSQL_TYPE_TIME:
		dt, err := unpackTime(br)
		return fmt.Sprintf("%s", dt), false, err
	}
	log.Printf("Unknown type = %s\n", f.Type)
	return "NULL", true, nil
}

func unpackDate(br *bufio.Reader) (dt *time.Time, err os.Error) {
	dt = new(time.Time)
	var y uint16
	var M, d, n uint8
	err = binary.Read(br, binary.LittleEndian, &n)
	if err != nil { return }
	err = binary.Read(br, binary.LittleEndian, &y)
	if err != nil { return }
	err = binary.Read(br, binary.LittleEndian, &M)
	if err != nil { return }
	err = binary.Read(br, binary.LittleEndian, &d)
	if err != nil { return }
	dt.Year = int64(y)
	dt.Month = int(M)
	dt.Day = int(d)
	dt.Hour = 0
	dt.Minute = 0
	dt.Second = 0
	return
}
func unpackTime(br *bufio.Reader) (dt *time.Time, err os.Error) {
	dt = new(time.Time)
	var h, m, s uint8
	err = ignoreBytes(br, 6)
	if err != nil { return }
	err = binary.Read(br, binary.LittleEndian, &h)
	if err != nil { return }
	err = binary.Read(br, binary.LittleEndian, &m)
	if err != nil { return }
	err = binary.Read(br, binary.LittleEndian, &s)
	if err != nil { return }
	dt.Year = 0
	dt.Month = 0
	dt.Day = 0
	dt.Hour = int(h)
	dt.Minute = int(m)
	dt.Second = int(s)
	return
}
func unpackDateTime(br *bufio.Reader) (dt *time.Time, err os.Error) {
	dt = new(time.Time)
	var y uint16
	var M, d, h, m, s, n uint8
	err = binary.Read(br, binary.LittleEndian, &n)
	if err != nil { return }
	err = binary.Read(br, binary.LittleEndian, &y)
	if err != nil { return }
	err = binary.Read(br, binary.LittleEndian, &M)
	if err != nil { return }
	err = binary.Read(br, binary.LittleEndian, &d)
	if err != nil { return }
	err = binary.Read(br, binary.LittleEndian, &h)
	if err != nil { return }
	err = binary.Read(br, binary.LittleEndian, &m)
	if err != nil { return }
	err = binary.Read(br, binary.LittleEndian, &s)
	if err != nil { return }
	dt.Year = int64(y)
	dt.Month = int(M)
	dt.Day = int(d)
	dt.Hour = int(h)
	dt.Minute = int(m)
	dt.Second = int(s)
	return
}


type MySQLField struct {
	Catalog  string
	Db       string
	Table    string
	OrgTable string
	Name     string
	OrgName  string

	Charset  uint16
	Length   uint32
	Type     uint8
	Flags    uint16
	Decimals uint8
	Default  uint64
}

func (f *MySQLField) String() string { return f.Name }

type MySQLData struct {
	Data   string
	Length uint64
	IsNull bool
	Type   uint8
}

func (d *MySQLData) String() string {
	if d.IsNull {
		return "NULL"
	}
	return string(d.Data)
}

type MySQLRow struct {
	Data []*MySQLData
}
