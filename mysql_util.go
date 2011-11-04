/*
	Utility functions for decoding mysql packets
*/

package mysql

import (
	"bufio"
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"fmt"
)

// Reads full slice or retutns error
func readFull(rd *bufio.Reader, p []byte) error {
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

//Read mysql packet header
func readHeader(br *bufio.Reader) (*PacketHeader, error) {
	ph := new(PacketHeader)
	i24seq := make([]byte, 4)
	err := readFull(br, i24seq)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("readHeader: %s", err))
	}
	ph.Len = unpackNumber(i24seq, 3)
	ph.Seq = i24seq[3]
	return ph, nil
}

//Decode length encoded number
func unpackLength(br *bufio.Reader) (val uint64, null bool, num int, err error) {
	var bl uint8
	err = binary.Read(br, binary.LittleEndian, &bl)
	if err != nil {
		return
	}
	num = 1
	if bl == 251 {
		null = true
	} else if bl == 252 {
		b := make([]byte, 2)
		num += 2
		err = readFull(br, b)
		val = unpackNumber(b, 2)
	} else if bl == 253 {
		b := make([]byte, 3)
		num += 3
		err = readFull(br, b)
		val = unpackNumber(b, 3)
	} else if bl == 254 {
		b := make([]byte, 8)
		num += 8
		err = readFull(br, b)
		val = unpackNumber(b, 8)
	} else {
		val = uint64(bl)
	}
	return
}

//Special case of unpackLength where 0xfe == EOF
func unpackFieldCount(br *bufio.Reader) (val uint64, null bool, err error) {
	var eof bool
	eof, err = peekEOF(br)
	if err != nil {
		return
	}
	if eof {
		err = ignoreBytes(br, 1)
		val = uint64(0xfe)
		return
	}
	val, null, _, err = unpackLength(br)
	return
}

func packUint8(bw *bufio.Writer, u uint8) error {
	return binary.Write(bw, binary.LittleEndian, u)
}

func packUint16(bw *bufio.Writer, u uint16) error {
	return binary.Write(bw, binary.LittleEndian, u)
}

func packUint24(bw *bufio.Writer, u uint32) error {
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

func packUint32(bw *bufio.Writer, u uint32) error {
	return binary.Write(bw, binary.LittleEndian, u)
}

func packUint64(bw *bufio.Writer, u uint64) error {
	return binary.Write(bw, binary.LittleEndian, u)
}

//Decode length encoded string
func unpackString(br *bufio.Reader) (string, bool, error) {
	length, isnull, _, err := unpackLength(br)
	if err != nil {
		return "", false, err
	}
	b := make([]byte, length)
	err = readFull(br, b)
	return string(b), isnull, err
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
func peekEOF(br *bufio.Reader) (bool, error) {
	b, err := br.ReadByte()
	if err != nil {
		return false, err
	}
	err = br.UnreadByte()
	if b == 0xfe {
		return true, err
	}
	return false, err
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
func readFieldPacket(br *bufio.Reader) (f *MySQLField, err error) {
	f = new(MySQLField)
	f.Catalog, _, err = unpackString(br)
	if err != nil {
		return
	}
	f.Db, _, err = unpackString(br)
	if err != nil {
		return
	}
	f.Table, _, err = unpackString(br)
	if err != nil {
		return
	}
	f.OrgTable, _, err = unpackString(br)
	if err != nil {
		return
	}
	f.Name, _, err = unpackString(br)
	if err != nil {
		return
	}
	f.OrgName, _, err = unpackString(br)
	if err != nil {
		return
	}
	err = ignoreBytes(br, 1)
	if err != nil {
		return
	}
	err = binary.Read(br, binary.LittleEndian, &f.Charset)
	if err != nil {
		return
	}
	err = binary.Read(br, binary.LittleEndian, &f.Length)
	if err != nil {
		return
	}
	err = binary.Read(br, binary.LittleEndian, &f.Type)
	if err != nil {
		return
	}
	err = binary.Read(br, binary.LittleEndian, &f.Flags)
	if err != nil {
		return
	}
	err = binary.Read(br, binary.LittleEndian, &f.Decimals)
	if err != nil {
		return
	}
	err = ignoreBytes(br, 1)
	if err != nil {
		return
	}
	f.Default, _, _, err = unpackLength(br)
	return
}

//Read error packet
func readErrorPacket(br *bufio.Reader, pkt_len int) error {
	var errcode uint16
	err := binary.Read(br, binary.LittleEndian, &errcode)
	if err != nil {
		return err
	}
	status := make([]byte, 6)
	err = readFull(br, status)
	if err != nil {
		return err
	}
	msg := make([]byte, pkt_len-9)
	err = readFull(br, msg)
	if err != nil {
		return err
	}
	return errors.New(fmt.Sprintf("MySQL Error: (Code: %d) (Status: %s) %s", errcode, string(status), string(msg)))
}

//Read EOF packet.
//TODO: Return something useful?
func readEOFPacket(br *bufio.Reader) (err error) {
	_, err = readHeader(br)
	if err != nil {
		return
	}
	response := new(MySQLResponse)
	response.FieldCount, _, err = unpackFieldCount(br)
	if err != nil {
		return
	}
	if response.FieldCount != 0xfe {
		fmt.Printf("Warning: Expected EOF! Got %#v\n", response.FieldCount)
	}
	err = binary.Read(br, binary.LittleEndian, &response.WarningCount)
	if err != nil {
		return
	}
	err = binary.Read(br, binary.LittleEndian, &response.ServerStatus)
	return
}

//Ignores n bytes in the buffer
func ignoreBytes(br *bufio.Reader, n uint64) error {
	buf := make([]byte, n)
	return readFull(br, buf)
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
