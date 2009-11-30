package mysql

import (
	"bufio";
	"encoding/binary";
	"os";
	"fmt";
	"reflect";
	"bytes";
	"strconv";
)

type MySQLStatement struct {
	StatementId	uint32;
	Columns		uint16;
	Parameters	uint16;
	Warnings	uint16;
	FieldCount	uint8;

	ResultSet	*MySQLResultSet;
	mysql		*MySQLInstance;
}

// Encode values fro each field
func encodeParamValues(a ...) ([]byte, int) {
	v := reflect.NewValue(a).(*reflect.StructValue);
	var b []byte;
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i);
		switch t := f.(type) {
		case *reflect.StringValue:
			b = bytes.Add(b, packString(string(t.Get())))
		case *reflect.IntValue:
			b = bytes.Add(b, packString(strconv.Itoa(int(t.Get()))))
		case *reflect.FloatValue:
			b = bytes.Add(b, packString(strconv.Ftoa(float(t.Get()), 'f', -1)))
		}
	}
	return b, len(b);
}

func putuint16(b []byte, v uint16) {
	b[0] = byte(v);
	b[1] = byte(v >> 8);
}

// For each field encode 2 byte type code. First bit is signed/unsigned
// Cheats and only bind parameters as strings
func encodeParamTypes(a ...) ([]byte, int) {
	v := reflect.NewValue(a).(*reflect.StructValue);
	buf := make([]byte, v.NumField()*2);
	off := 0;
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i);
		if reflect.Indirect(f) == nil {
			putuint16(buf[off:off+2], uint16(MYSQL_TYPE_NULL));
			continue;
		}
		putuint16(buf[off:off+2], uint16(MYSQL_TYPE_STRING));
		off += 2;
	}
	return buf, off;
}


func readPrepareInit(br *bufio.Reader) (*MySQLStatement, os.Error) {
	ph := readHeader(br);
	s := new(MySQLStatement);
	err := binary.Read(br, binary.LittleEndian, &s.FieldCount);
	if s.FieldCount == uint8(0xff) {
		return nil, readErrorPacket(br);
	}
	err = binary.Read(br, binary.LittleEndian, &s.StatementId);
	err = binary.Read(br, binary.LittleEndian, &s.Columns);
	err = binary.Read(br, binary.LittleEndian, &s.Parameters);
	if ph.Len >= 12 {
		ignoreBytes(br, 1);
		err = binary.Read(br, binary.LittleEndian, &s.Warnings);
	}
	return s, err;
}

//Currently just skips the pakets as I'm not sure if they are useful.
func readPrepareParameters(br *bufio.Reader, s *MySQLStatement) os.Error {
	for i := uint16(0); i < s.Parameters; i++ {
		ph := readHeader(br);
		ignoreBytes(br, int(ph.Len));
	}
	return nil;
}

func (sth *MySQLStatement) execute(va ...) (*MySQLResponse, os.Error) {
	v := reflect.NewValue(va).(*reflect.StructValue);
	if int(sth.Parameters) != v.NumField() {
		return nil, os.ErrorString(fmt.Sprintf("Parameter count mismatch. %d != %d", sth.Parameters, v.NumField()))
	}
	type_parm, tn := encodeParamTypes(va);
	value_parm, vn := encodeParamValues(va);
	bitmap_len := (v.NumField() + 7) / 8;
	mysql := sth.mysql;
	packUint24(mysql.writer, uint32(11+bitmap_len+tn+vn));
	packUint8(mysql.writer, uint8(0));
	packUint8(mysql.writer, uint8(COM_STMT_EXECUTE));
	packUint32(mysql.writer, uint32(sth.StatementId));
	packUint8(mysql.writer, uint8(0));
	packUint32(mysql.writer, uint32(1));
	b := make([]byte, bitmap_len);
	mysql.writer.Write(b);	//TODO: Support null params.
	packUint8(mysql.writer, uint8(1));
	mysql.writer.Write(type_parm);
	mysql.writer.Write(value_parm);
	mysql.writer.Flush();
	res, err := mysql.readResult();
	res.Prepared = true;
	return res, err;
}

func (mysql *MySQLInstance) prepare(arg string) (*MySQLStatement, os.Error) {
	plen := len(arg) + 1;
	var head [5]byte;
	head[0] = byte(plen);
	head[1] = byte(plen >> 8);
	head[2] = byte(plen >> 16);
	head[3] = 0;
	head[4] = uint8(COM_STMT_PREPARE);
	_, err := mysql.writer.Write(&head);
	err = mysql.writer.WriteString(arg);
	err = mysql.writer.Flush();
	if err != nil {
		fmt.Printf("%s\n", err);
		return nil, err;
	}
	sth, err := readPrepareInit(mysql.reader);
	if err != nil {
		return nil, err;
	}
	if sth.Parameters > 0 {
		readPrepareParameters(mysql.reader, sth)
	}
	readEOFPacket(mysql.reader);
	if sth.Columns > 0 {
		rs, _ := mysql.readResultSet(uint64(sth.Columns));
		sth.ResultSet = rs;
	}
	sth.mysql = mysql;
	return sth, nil;
}
