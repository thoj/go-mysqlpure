package mysql

import (
	"encoding/binary";
	"bufio";
)

// Read Length Encoded Binary
func readLengthCodedBinary(br *bufio.Reader) (uint8, []byte) {
	var bl uint8;
	binary.Read(br, binary.LittleEndian, &bl);
	var b []byte = make([]byte, bl);
	br.Read(b);
	return bl, b;
}

//Convert n bytes to uint64 (Little Endian)
func byteToUIntLE(b []byte, n uint8) uint64 {
	if n < 1 {
		return 0
	}
	var r uint64 = 0;
	for i := uint8(0); i < n; i++ {
		r += uint64(b[i] << (i * 8))
	}
	return r;
}
