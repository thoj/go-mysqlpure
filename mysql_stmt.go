package mysql

import (
	"bufio";
	"encoding/binary";
	"os";
	"fmt";
	"reflect";
)

type MySQLStatement struct {
	StatementId	uint32;
	Columns		uint16;
	Parameters	uint16;
	Warnings	uint16;

	ResultSet	*MySQLResultSet;
	mysql		*MySQLInstance;
}

func encodeTypes(a ...) {
        v := reflect.NewValue(a).(*reflect.StructValue);
        fmt.Printf("%d\n", v.NumField());
        for i := 0; i < v.NumField(); i ++ {
                f := v.Field(i);
                fmt.Printf("%#v\n", f);
                if _, ok := f.(*reflect.StringValue); ok {
                        fmt.Printf("%s\n", "STRING");
                }
        }
}


func (sth *MySQLStatement) Execute(va ...) os.Error {
	if sth.Parameters != len(va) {
		return os.ErrorString(fmt.Sprintf("Parameter count mismatch. %d != %d", sth.Parameters, len(va)));
	}
	bitmap_len = (len(va) + 7)/8;
	mysql := sth.mysql;
	packUint24(mysql.writer, uint32(11+bitmap_len+(len(va))*2));
	packUint8(mysql.writer, uint8(1));
	packUint8(mysql.writer, uint8(COM_STMT_EXECUTE));
	packUint32(mysql.writer, uint32(sth.StatementId));
	packUint8(mysql.writer, uint8(0));
	packUint32(mysql.writer, uint32(1));
	b := make([]byte, bitmap_len);
	mysql.writer.Write(b); //TODO: Support null params.
	packUint8(mysql.writer, uint8(1));
	encodeTypes(va);}
}

func readPrepareInit(br *bufio.Reader) (*MySQLStatement, os.Error) {
	ph := readHeader(br);
	s := new(MySQLStatement);
	ignoreBytes(br, 1);
	err := binary.Read(br, binary.LittleEndian, &s.StatementId);
	err = binary.Read(br, binary.LittleEndian, &s.Columns);
	err = binary.Read(br, binary.LittleEndian, &s.Parameters);
	if ph.Len >= 12 {
		ignoreBytes(br, 1);
		err = binary.Read(br, binary.LittleEndian, &s.Warnings);
//		fmt.Printf("Warnings = %x\n", s.Warnings);
	}
	return s, err;
}

//Currently just skips the pakets as I'm not sure if they are useful.
func readPrepareParameters(br *bufio.Reader, s *MySQLStatement) os.Error {
	for i := uint16(0); i < s.Parameters; i++ {
		ph := readHeader(br);
		ignoreBytes(br, int(ph.Len));
//		fmt.Printf("Ignoring %d bytes\n", ph.Len);
	}
	return nil;
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
//	fmt.Printf("%#v\n", sth);
	if sth.Parameters > 0 {
		readPrepareParameters(mysql.reader, sth)
	}
	readEOFPacket(mysql.reader);
	if sth.Columns > 0 {
		rs, err := mysql.readResultSet(uint64(sth.Columns));
		sth.ResultSet = rs;
	}
	readEOFPacket(mysql.reader);
	sth.mysql = mysql;
	return sth, nil;
}
