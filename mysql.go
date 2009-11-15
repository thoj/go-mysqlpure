// Copyright 2009 Thomas Jager <mail@jager.no>  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// MySQL 4.1+ Client Library.

package mysql


import (
	"fmt";
	"net";
	"os";
	"bytes";
	"bufio";
	"encoding/binary";
	"crypto/sha1";
	"strings";
)


// Constants
const (
	CLIENT_LONG_PASSWORD		= 1;		/* new more secure passwords */
	CLIENT_FOUND_ROWS		= 2;		/* Found instead of affected rows */
	CLIENT_LONG_FLAG		= 4;		/* Get all column flags */
	CLIENT_CONNECT_WITH_DB		= 8;		/* One can specify db on connect */
	CLIENT_NO_SCHEMA		= 16;		/* Don't allow database.table.column */
	CLIENT_COMPRESS			= 32;		/* Can use compression protocol */
	CLIENT_ODBC			= 64;		/* Odbc client */
	CLIENT_LOCAL_FILES		= 128;		/* Can use LOAD DATA LOCAL */
	CLIENT_IGNORE_SPACE		= 256;		/* Ignore spaces before '(' */
	CLIENT_PROTOCOL_41		= 512;		/* New 4.1 protocol */
	CLIENT_INTERACTIVE		= 1024;		/* This is an interactive client */
	CLIENT_SSL			= 2048;		/* Switch to SSL after handshake */
	CLIENT_IGNORE_SIGPIPE		= 4096;		/* IGNORE sigpipes */
	CLIENT_TRANSACTIONS		= 8192;		/* Client knows about transactions */
	CLIENT_RESERVED			= 16384;	/* Old flag for 4.1 protocol  */
	CLIENT_SECURE_CONNECTION	= 32768;	/* New 4.1 authentication */
	CLIENT_MULTI_STATEMENTS		= 65536;	/* Enable/disable multi-stmt support */
	CLIENT_MULTI_RESULTS		= 131072;	/* Enable/disable multi-results */
)

//Common Header
type PacketHeaderR struct {
	Len1	uint8;
	Len2	uint8;
	Len3	uint8;
	Seq	uint8;
}

var ProtocolVersion uint8	// Protocol version = 0x10
var ServerVersion string	// Server string
var ThreadId uint32		// Current Thread ID
var ServerCapabilities uint16
var ServerLanguage uint8
var ServerStatus uint16

var Connected = false


var scrambleBuffer []byte

//Read initial handshake packet.
func readInit(br *bufio.Reader) os.Error {
	var ph PacketHeaderR;
	binary.Read(br, binary.LittleEndian, &ph);	//Header (Length and Seqence number)
	if ph.Seq != 0 {
		// Initial packet must be Seq == 0
		return os.ErrorString("Unexpected Sequence Number")
	}
	binary.Read(br, binary.LittleEndian, &ProtocolVersion);
	ServerVersion, _ = br.ReadString('\x00');
	binary.Read(br, binary.LittleEndian, &ThreadId);
	var sb [9]byte;
	br.Read(&sb);
	binary.Read(br, binary.LittleEndian, &ServerCapabilities);
	binary.Read(br, binary.LittleEndian, &ServerLanguage);
	binary.Read(br, binary.LittleEndian, &ServerStatus);
	var sb2 [26]byte;
	br.Read(&sb2);
	scrambleBuffer = new([20]byte);
	bytes.Copy(scrambleBuffer[0:8], sb[0:8]);
	bytes.Copy(scrambleBuffer[8:20], sb2[13:25]);
	return nil;
}

//Tries to read OK result error on error packett
func readResult(br *bufio.Reader) os.Error {
	var ph PacketHeaderR;
	err := binary.Read(br, binary.LittleEndian, &ph);
	var result byte;
	err = binary.Read(br, binary.LittleEndian, &result);
	if result == 0xff {
		var errcode uint16;
		binary.Read(br, binary.LittleEndian, &errcode);
		status := make([]byte, 6);
		br.Read(status);
		msg, _ := br.ReadString(0x00);
		return os.ErrorString(fmt.Sprintf("MySQL Error: (Code: %d) (Status: %s) %s", errcode, string(status), msg));
	}
	if err != nil {
		return err
	}
	fmt.Printf("Result == %x\n", result);
	return nil;
}

//Connects to mysql server and reads the initial handshake,
//then tries to login using supplied credentials.
func Connect(host string, username string, password string, database string) os.Error {
	conn, err := net.Dial("tcp", "", host);
	if err != nil {
		return os.ErrorString(fmt.Sprintf("Cant connect to %s\n", host))
	}
	br := bufio.NewReader(conn);
	bw := bufio.NewWriter(conn);
	if err = readInit(br); err != nil {
		return err
	}
	err = sendAuth(bw, database, username, password);
	if err = readResult(br); err != nil {
		return err
	}
	Connected = true;
	fmt.Printf("Connected to server\n");
	return nil;
}

//This is really ugly.
func mysqlPassword(password []byte) []byte {
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

// Try to auth using the MySQL secure auth *crossing fingers*
func sendAuth(bw *bufio.Writer, database string, username string, password string) os.Error {
	var clientattr uint32 = CLIENT_LONG_PASSWORD + CLIENT_PROTOCOL_41 + CLIENT_SECURE_CONNECTION;
	var plen int = len(username);
	if len(database) > 0 {
		clientattr += CLIENT_CONNECT_WITH_DB;
		plen += len(database) + 55
	} else {
		plen += 54
	}
	var head [13]byte;
	head[0] = byte(plen);
	head[1] = byte(plen >> 8);
	head[2] = byte(plen >> 16);
	head[3] = 1;
	binary.LittleEndian.PutUint32(head[4:8], clientattr);
	binary.LittleEndian.PutUint32(head[8:12], uint32(1073741824));
	head[12] = ServerLanguage;
	bw.Write(&head);
	var filler [23]byte;
	bw.Write(&filler);
	bw.WriteString(username);
	bw.Write(filler[0:1]);
	token := mysqlPassword(strings.Bytes(password));
	bw.Write(token);
	if len(database) > 0 {
		bw.WriteString(database);
		bw.Write(filler[0:1]);
	}
	bw.Flush();

	return nil;

}
