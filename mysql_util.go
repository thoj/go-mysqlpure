package mysql

import (
	"encoding/binary";
	"bufio";
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
