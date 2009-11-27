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

// Encode values fro each field
func encodeParamValues(bw *bufio.Writer, a ...) {
	v := reflect.NewValue(a).(*reflect.StructValue);
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i);
		fmt.Printf("%#v\n", f);
		switch f.(type) {
		case *reflect.StringValue:
			fmt.Printf("STRING\n")
		case *reflect.IntValue:
			fmt.Printf("INT\n")
		}
	}
}

// For each field encode 2 byte type code. First bit is signed/unsigned
func encodeParamTypes(bw *bufio.Writer, a ...) {
	v := reflect.NewValue(a).(*reflect.StructValue);
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i);
		fmt.Printf("%#v\n", f);
		switch f.(type) {
		case *reflect.StringValue:
			fmt.Printf("STRING\n")
		case *reflect.IntValue:
			fmt.Printf("INT\n")
		}
	}
}

func (sth *MySQLStatement) Execute(va ...) os.Error {
	v := reflect.NewValue(va).(*reflect.StructValue);
	if int(sth.Parameters) != v.NumField() {
		return os.ErrorString(fmt.Sprintf("Parameter count mismatch. %d != %d", sth.Parameters, v.NumField()))
	}
	bitmap_len := (v.NumField() + 7) / 8;
	mysql := sth.mysql;
	packUint24(mysql.writer, uint32(11+bitmap_len+v.NumField()*2));
	packUint8(mysql.writer, uint8(1));
	packUint8(mysql.writer, uint8(COM_STMT_EXECUTE));
	packUint32(mysql.writer, uint32(sth.StatementId));
	packUint8(mysql.writer, uint8(0));
	packUint32(mysql.writer, uint32(1));
	b := make([]byte, bitmap_len);
	mysql.writer.Write(b);	//TODO: Support null params.
	packUint8(mysql.writer, uint8(1));
	fmt.Printf("%d\n", v.NumField());
	encodeParamTypes(mysql.writer, va);
	encodeParamValues(mysql.writer, va);
	return nil;
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
		rs, _ := mysql.readResultSet(uint64(sth.Columns));
		sth.ResultSet = rs;
	}
	readEOFPacket(mysql.reader);
	sth.mysql = mysql;
	return sth, nil;
}
