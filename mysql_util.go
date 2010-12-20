/*
	Utility functions for decoding mysql packets
*/

package mysql

import (
	"encoding/binary"
	"bufio"
	"os"
	"fmt"
	"crypto/sha1"
)

// Reads full slice or retutns error
func readFull(rd *bufio.Reader, p []byte) os.Error {
    for nn := 0; nn < len(p); {
        kk, err := rd.Read(p[nn:])
        /*if kk != len(p) {
            fmt.Printf("DEBUG: readed %d/%d\n", kk, len(p))
        }*/
        if err != nil {
            return err
        }
        nn += kk
    }
    return nil
}

func mustReadFull(rd *bufio.Reader, p []byte) {
    err := readFull(rd, p)
    if err != nil {
        panic(err)
    }
}

//Read mysql packet header
func readHeader(br *bufio.Reader) (*PacketHeader, os.Error) {
	ph := new(PacketHeader)
	i24seq := make([]byte, 4)
        err := readFull(br, i24seq)
	if err != nil {
		return nil, os.ErrorString(fmt.Sprintf("readHeader: %s", err))
	}
	ph.Len = unpackNumber(i24seq, 3)
	ph.Seq = i24seq[3]
	return ph, nil
}

//Decode length encoded number
//TODO: *Decode() Check Buffered bytes.
func unpackLength(br *bufio.Reader) (uint64, bool) {
	var bl uint8
	binary.Read(br, binary.LittleEndian, &bl)
	if bl < 251 {
		return uint64(bl), false
	} else if bl == 251 {
		return 0, true
	} else if bl == 252 {
		b := make([]byte, 2)
		mustReadFull(br, b)
		return unpackNumber(b, 2), false
	} else if bl == 253 {
		b := make([]byte, 3)
		mustReadFull(br, b)
		return unpackNumber(b, 3), false
	} else if bl == 254 && br.Buffered() > 8 {
		b := make([]byte, 8)
		mustReadFull(br, b)
		return unpackNumber(b, 8), false
	}
	return uint64(bl), false
}

//Special case of unpackLength where 0xfe == EOF
func unpackFieldCount(br *bufio.Reader) (uint64, bool) {
	if peekEOF(br) {
		ignoreBytes(br, 1)
		return uint64(0xfe), false
	}
	return unpackLength(br)
}

func packUint8(bw *bufio.Writer, u uint8) os.Error {
	return binary.Write(bw, binary.LittleEndian, u)
}

func packUint16(bw *bufio.Writer, u uint16) os.Error {
	return binary.Write(bw, binary.LittleEndian, u)
}

func packUint24(bw *bufio.Writer, u uint32) os.Error {
	b := make([]byte, 3)
	b[0] = byte(u)
	b[1] = byte(u >> 8)
	b[2] = byte(u >> 16)
	n, err := bw.Write(b)
	if n != 3 {
		return err
	}
	return nil
}

func packUint32(bw *bufio.Writer, u uint32) os.Error {
	return binary.Write(bw, binary.LittleEndian, u)
}

func packUint64(bw *bufio.Writer, u uint64) os.Error {
	return binary.Write(bw, binary.LittleEndian, u)
}

//Decode length encoded string
func unpackString(br *bufio.Reader) (string, bool) {
	length, isnull := unpackLength(br)
	b := make([]byte, length)
	mustReadFull(br, b)
	return string(b), isnull
}

func packString(s string) []byte {
	sb := []byte(s)
	size := make([]byte, 1)
	v := len(sb)
	if v < 250 {
		size[0] = uint8(v)
		return append(size, sb...)
	}
	size = make([]byte, 9)
	size[0] = 254
	size[1] = byte(v)
	size[2] = byte(v >> 8)
	size[3] = byte(v >> 16)
	size[4] = byte(v >> 24)
	size[5] = byte(v >> 32)
	size[6] = byte(v >> 40)
	size[7] = byte(v >> 48)
	size[8] = byte(v >> 56)
	return append(size, sb...)
}

//Peek and check if packet is EOF
func peekEOF(br *bufio.Reader) bool {
	b, err := br.ReadByte()
        if err != nil {
                panic(err)
        }
	br.UnreadByte()
	if b == 0xfe {
		return true
	}
	return false
}

//Convert n bytes to uint64 (Little Endian)
func unpackNumber(b []byte, n uint8) uint64 {
	if n < 1 {
		return 0
	}
	var r uint64 = 0
	for i := uint8(0); i < n; i++ {
		r |= (uint64(b[i]) << (i * 8))
	}
	return r
}

//Read the field data from mysql ResultSet packet
func readFieldPacket(br *bufio.Reader) *MySQLField {
	f := new(MySQLField)
	f.Catalog, _ = unpackString(br)
	f.Db, _ = unpackString(br)
	f.Table, _ = unpackString(br)
	f.OrgTable, _ = unpackString(br)
	f.Name, _ = unpackString(br)
	f.OrgName, _ = unpackString(br)
	ignoreBytes(br, 1)
	binary.Read(br, binary.LittleEndian, &f.Charset)
	binary.Read(br, binary.LittleEndian, &f.Length)
	binary.Read(br, binary.LittleEndian, &f.Type)
	binary.Read(br, binary.LittleEndian, &f.Flags)
	binary.Read(br, binary.LittleEndian, &f.Decimals)
	ignoreBytes(br, 1)
	eb, _ := unpackLength(br)
	f.Default = eb
	return f
}

//Read error packet
func readErrorPacket(br *bufio.Reader) os.Error {
	var errcode uint16
	binary.Read(br, binary.LittleEndian, &errcode)
	status := make([]byte, 6)
	err := readFull(br, status)
        if err != nil {
            return err
        }
	msg := make([]byte, br.Buffered())
	err = readFull(br, msg)
        if err != nil {
            return err
        }
	return os.ErrorString(fmt.Sprintf("MySQL Error: (Code: %d) (Status: %s) %s", errcode, string(status), string(msg)))
}

//Read EOF packet.
//TODO: Return something useful?
func readEOFPacket(br *bufio.Reader) os.Error {
	readHeader(br)

	response := new(MySQLResponse)
	response.FieldCount, _ = unpackFieldCount(br)
	if response.FieldCount != 0xfe {
		fmt.Printf("Warning: Expected EOF! Got %#v\n", response.FieldCount)
	}
	binary.Read(br, binary.LittleEndian, &response.WarningCount)
	binary.Read(br, binary.LittleEndian, &response.ServerStatus)
	return nil
}

//Ignores n bytes in the buffer
func ignoreBytes(br *bufio.Reader, n uint64) {
	buf := make([]byte, n)
	mustReadFull(br, buf)
}

//Generate scrabled password using password and scramble buffer.
func mysqlPassword(password []byte, scrambleBuffer []byte) []byte {
	ctx := sha1.New()
	ctx.Write(password)
	stage1 := ctx.Sum()

	ctx.Reset()
	ctx.Write(stage1)
	stage2 := ctx.Sum()

	ctx.Reset()
	ctx.Write(scrambleBuffer)
	ctx.Write(stage2)
	result := ctx.Sum()

	token := make([]byte, 21)
	token_t := make([]byte, 21)
	for i := 0; i < 20; i++ {
		token[i+1] = result[i] ^ stage1[i]
	}
	for i := 0; i < 20; i++ {
		token_t[i] = token[i+1] ^ result[i]
	}
	token[0] = 20
	return token
}
