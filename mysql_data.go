package mysql

import (
	"fmt";
)

type PacketHeader struct {
	Len	uint64;
	Seq	uint8;
}

type MySQLResultSet struct {
	FieldCount	uint64;
	Fields		[]*MySQLField;
	Rows		[]*MySQLRow;
}

type MySQLResponse struct {
	FieldCount	uint64;
	AffectedRows	uint64;
	InsertId	uint64;
	ServerStatus	uint16;
	WarningCount	uint16;
	Message		[]string;
	EOF		bool;
	Prepared 	bool; //Result from prapered statement

	ResultSet	*MySQLResultSet;
	mysql		*MySQLInstance;
}


func (s *MySQLStatement) String() string {
	return fmt.Sprintf("Statement Id = %d, Columns = %d, Parameters = %d", s.StatementId, s.Columns, s.Parameters)
}

func (r *MySQLResponse) String() string {
	var msg string;
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
	msg = fmt.Sprintf("%s, Server Status = %x", msg, r.ServerStatus);
	if r.WarningCount > 0 {
		msg = fmt.Sprintf("%s, Warnings = %x", msg, r.WarningCount)
	}
	if len(r.Message) > 0 {
		msg = fmt.Sprintf("%s, Message = %s", msg, r.Message)
	}
	return msg;
}

type MySQLField struct {
	Catalog		string;
	Db		string;
	Table		string;
	OrgTable	string;
	Name		string;
	OrgName		string;

	Charset		uint16;
	Length		uint32;
	Type		uint8;
	Flags		uint16;
	Decimals	uint8;
	Default		uint64;
}

func (f *MySQLField) String() string	{ return f.Name }

type MySQLData struct {
	Data	string;
	Length	uint64;
	IsNull	bool;
	Type	uint8;
}

func (d *MySQLData) String() string {
	if d.IsNull {
		return "NULL"
	}
	return string(d.Data);
}

type MySQLRow struct {
	Data []*MySQLData;
}
