package mysql

import (
	"encoding/binary";
	"bufio";
	"os";
	"fmt";
	"crypto/sha1";
)


type PacketHeader struct {
	Len	uint64;
	Seq	uint8;
}


func readHeader(br *bufio.Reader) *PacketHeader {
	ph := new(PacketHeader);
	var i24seq [4]byte;
	br.Read(&i24seq);
	ph.Len = unpackNumber(&i24seq, 3);
	ph.Seq = i24seq[3];
	return ph;
}

func unpackLength(br *bufio.Reader) (uint64, bool) {
	var bl uint8;
	binary.Read(br, binary.LittleEndian, &bl);
	if bl < 251 {
		return uint64(bl), false
	} else if bl == 251 {
		return 0, true
	} else if bl == 252 {
		b := make([]byte, 2);
		br.Read(b);
		return unpackNumber(b, 2), false;
	} else if bl == 253 {
		b := make([]byte, 3);
		br.Read(b);
		return unpackNumber(b, 3), false;
	}
	b := make([]byte, 8);
	n, _ := br.Read(b);
	if n  == 8 {
		return unpackNumber(b, 8), false;
	} 
	return 0, true;
}

func unpackString(br *bufio.Reader) (string, bool) {
	length, isnull := unpackLength(br);
	b := make([]byte, length);
	br.Read(b);
	return string(b), isnull;
}

func peekEOF(br *bufio.Reader) bool {
	b := make([]byte, 1);
	br.Read(b);
	br.UnreadByte();
	if b[0] == 0xfe {
		return true;
	}
	return false;
}
//Convert n bytes to uint64 (Little Endian)
func unpackNumber(b []byte, n uint8) uint64 {
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
	f.Catalog, _ = unpackString(br);
	f.Db, _ = unpackString(br);
	f.Table, _ = unpackString(br);
	f.OrgTable, _ = unpackString(br);
	f.Name, _ = unpackString(br);
	f.OrgName, _ = unpackString(br);
	var filler [2]byte;
	br.Read(filler[0:1]);
	binary.Read(br, binary.LittleEndian, &f.Charset);
	binary.Read(br, binary.LittleEndian, &f.Length);
	binary.Read(br, binary.LittleEndian, &f.Type);
	binary.Read(br, binary.LittleEndian, &f.Flags);
	binary.Read(br, binary.LittleEndian, &f.Decimals);
	br.Read(filler[0:1]);
	eb,_ := unpackLength(br);
	f.Default = eb;
  	fmt.Printf("%#v\n", f);	
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
