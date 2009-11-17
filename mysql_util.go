package mysql

import (
	"encoding/binary";
	"bufio";
	"os";
	"fmt";
	"crypto/sha1";	
)


type EncodedBinary struct {
	Length	uint8;
	IsNull	bool;
	Value	uint64;
}

type PacketHeader struct {
	Len	uint64;
	Seq	uint8;
}


func readHeader(br *bufio.Reader) *PacketHeader {
	ph := new(PacketHeader);
	var i24seq [4]byte;
	br.Read(&i24seq);
	ph.Len = byteToUInt64LE(&i24seq, 3);
	ph.Seq = i24seq[3];
	return ph;
}
func readLengthCodedString(br *bufio.Reader) string {
	var bl uint8;
	binary.Read(br, binary.LittleEndian, &bl);
	b := make([]byte,bl);
	br.Read(b);
	return string(b);	
}

// Read Length Encoded Binary
func readLengthCodedBinary(br *bufio.Reader) *EncodedBinary {
	var bl uint8;
	var eb EncodedBinary;
	binary.Read(br, binary.LittleEndian, &bl);
	if bl >= 0 && bl < 251 {
		eb.Length = 1;
		eb.Value = uint64(bl);
		eb.IsNull = false;
		return &eb;
	}
	switch bl {
	case 251:
		eb.IsNull = true;
		eb.Length = 0;
		eb.Value = 0;
	case 252:
		eb.IsNull = false;
		eb.Length = 2;
		var i16 uint16;
		binary.Read(br, binary.LittleEndian, &i16);
		eb.Value = uint64(i16);
	case 253:
		eb.IsNull = false;
		eb.Length = 3;
		var i24 [3]byte;
		br.Read(&i24);
		eb.Value = byteToUInt64LE(&i24, 3);
	case 254:
		eb.IsNull = false;
		eb.Length = 2;
		var i64 uint64;
		binary.Read(br, binary.LittleEndian, &i64);
		eb.Value = i64;
		;
	}
	return &eb;
}

//Convert n bytes to uint64 (Little Endian)
func byteToUInt64LE(b []byte, n uint8) uint64 {
	if n < 1 {
		return 0
	}
	var r uint64 = 0;
	for i := uint8(0); i < n; i++ {
		r += uint64(b[i] << (i * 8))
	}
	return r;
}

//Convert n bytes to uint32 (Little Endian)
func byteToUInt32LE(b []byte, n uint8) uint32 {
	if n < 1 {
		return 0
	}
	var r uint32 = 0;
	for i := uint8(0); i < n; i++ {
		r += uint32(b[i] << (i * 8))
	}
	return r;
}

func readFieldPacket(br *bufio.Reader) *MySQLField {
	f := new(MySQLField);
	f.Catalog = readLengthCodedString(br);
	f.Db = readLengthCodedString(br);
	f.Table = readLengthCodedString(br);
	f.OrgTable = readLengthCodedString(br);
	f.Name = readLengthCodedString(br);
	f.OrgName = readLengthCodedString(br);
	var filler [2]byte;
	br.Read(filler[0:1]);
	binary.Read(br, binary.LittleEndian, &f.Charset);
	binary.Read(br, binary.LittleEndian, &f.Length);
	binary.Read(br, binary.LittleEndian, &f.Type);
	binary.Read(br, binary.LittleEndian, &f.Flags);
	binary.Read(br, binary.LittleEndian, &f.Decimals);
	br.Read(filler[0:1]);
	eb := readLengthCodedBinary(br);
	f.Default = eb.Value;
	return f;
}

func readEOFPacket(br *bufio.Reader) os.Error {
	readHeader(br);

	response := new(MySQLResponse);
	binary.Read(br, binary.LittleEndian, &response.FieldCount);
	if response.FieldCount != 0xfe {
		fmt.Printf("Expected EOF! Got %#v\n", response.FieldCount)
	}
	binary.Read(br, binary.LittleEndian, &response.WarningCount);
	binary.Read(br, binary.LittleEndian, &response.ServerStatus);
	return nil;
}


//This is really ugly.
func mysqlPassword(password []byte, scrambleBuffer []byte) []byte {
	ctx := sha1.New();
	ctx.Write(password);
	stage1 := ctx.Sum();

	ctx = sha1.New();
	ctx.Write(stage1);
	stage2 := ctx.Sum();

	ctx = sha1.New();
	ctx.Write(scrambleBuffer);
	ctx.Write(stage2);
	result := ctx.Sum();

	token := new([21]byte);
	token_t := new([20]byte);
	for i := 0; i < 20; i++ {
		token[i+1] = result[i] ^ stage1[i]
	}
	for i := 0; i < 20; i++ {
		token_t[i] = token[i+1] ^ result[i]
	}
	token[0] = 20;
	return token;
}
