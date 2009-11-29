/*
	Utility functions for decoding mysql packets
*/

package mysql

import (
	"encoding/binary";
	"bufio";
	"os";
	"fmt";
	"crypto/sha1";
	"strings";
	"bytes";
)

//Read mysql packet header
func readHeader(br *bufio.Reader) *PacketHeader {
	ph := new(PacketHeader);
	var i24seq [4]byte;
	br.Read(&i24seq);
	ph.Len = unpackNumber(&i24seq, 3);
	ph.Seq = i24seq[3];
	fmt.Printf("%s\n", ph);
	return ph;
}

//Decode length encoded number
//TODO: *Decode() Check Buffered bytes.
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
	} else if bl == 254 && br.Buffered() > 8 {
		b := make([]byte, 8);
		br.Read(b);
		return unpackNumber(b, 8), false;
	}
	return uint64(bl), false;
}

//Special case of unpackLength where 0xfe == EOF
func unpackFieldCount(br *bufio.Reader) (uint64, bool) {
	if peekEOF(br) {
		ignoreBytes(br, 1);
		return uint64(0xfe), false;
	}
	return unpackLength(br);
}

func packUint8(bw *bufio.Writer, u uint8) os.Error {
	return binary.Write(bw, binary.LittleEndian, u)
}

func packUint16(bw *bufio.Writer, u uint16) os.Error {
	return binary.Write(bw, binary.LittleEndian, u)
}

func packUint24(bw *bufio.Writer, u uint32) os.Error {
	b := make([]byte, 3);
	b[0] = byte(u);
	b[1] = byte(u >> 8);
	b[2] = byte(u >> 16);
	n, err := bw.Write(b);
	if n != 3 {
		return err
	}
	return nil;
}

func packUint32(bw *bufio.Writer, u uint32) os.Error {
	return binary.Write(bw, binary.LittleEndian, u)
}

func packUint64(bw *bufio.Writer, u uint64) os.Error {
	return binary.Write(bw, binary.LittleEndian, u)
}

//Decode length encoded string
func unpackString(br *bufio.Reader) (string, bool) {
	length, isnull := unpackLength(br);
	b := make([]byte, length);
	br.Read(b);
	return string(b), isnull;
}

func packString(s string) []byte {
	sb := strings.Bytes(s);
	size := make([]byte, 1);
	v := len(sb);
	if v < 250 {
		size[0] = uint8(v);
		return bytes.Add(size, sb);
	}
	size = make([]byte, 9);
	size[0] = 254;
	size[1] = byte(v);
	size[2] = byte(v >> 8);
	size[3] = byte(v >> 16);
	size[4] = byte(v >> 24);
	size[5] = byte(v >> 32);
	size[6] = byte(v >> 40);
	size[7] = byte(v >> 48);
	size[8] = byte(v >> 56);
	return bytes.Add(size, sb);
}

//Peek and check if packet is EOF
func peekEOF(br *bufio.Reader) bool {
	b := make([]byte, 1);
	br.Read(b);
	br.UnreadByte();
	if b[0] == 0xfe {
		return true
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
		r |= (uint64(b[i]) << (i * 8))
	}
	return r;
}

//Read the field data from mysql ResultSet packet
func readFieldPacket(br *bufio.Reader) *MySQLField {
	f := new(MySQLField);
	f.Catalog, _ = unpackString(br);
	f.Db, _ = unpackString(br);
	f.Table, _ = unpackString(br);
	f.OrgTable, _ = unpackString(br);
	f.Name, _ = unpackString(br);
	f.OrgName, _ = unpackString(br);
	ignoreBytes(br, 1);
	binary.Read(br, binary.LittleEndian, &f.Charset);
	binary.Read(br, binary.LittleEndian, &f.Length);
	binary.Read(br, binary.LittleEndian, &f.Type);
	binary.Read(br, binary.LittleEndian, &f.Flags);
	binary.Read(br, binary.LittleEndian, &f.Decimals);
	ignoreBytes(br, 1);
	eb, _ := unpackLength(br);
	f.Default = eb;
	return f;
}

//Read EOF packet.
//TODO: Return something useful?
func readEOFPacket(br *bufio.Reader) os.Error {
	readHeader(br);

	response := new(MySQLResponse);
	response.FieldCount, _ = unpackFieldCount(br);
	if response.FieldCount != 0xfe {
		fmt.Printf("Expected EOF! Got %#v\n", response.FieldCount)
	}
	binary.Read(br, binary.LittleEndian, &response.WarningCount);
	binary.Read(br, binary.LittleEndian, &response.ServerStatus);
	return nil;
}

//Ignores n bytes in the buffer
func ignoreBytes(br *bufio.Reader, n int) {
	buf := make([]byte, n);
	br.Read(buf);
}

//Generate scrabled password using password and scramble buffer.
func mysqlPassword(password []byte, scrambleBuffer []byte) []byte {
	ctx := sha1.New();
	ctx.Write(password);
	stage1 := ctx.Sum();

	ctx.Reset();
	ctx.Write(stage1);
	stage2 := ctx.Sum();

	ctx.Reset();
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
