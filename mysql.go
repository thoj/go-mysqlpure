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

//Common Header
type PacketHeaderR struct {
	Len1	uint8;
	Len2	uint8;
	Len3	uint8;
	Seq	uint8;
}

type MySQLInstance struct {
	ProtocolVersion		uint8;	// Protocol version = 0x10
	ServerVersion		string;	// Server string
	ThreadId		uint32;	// Current Thread ID
	ServerCapabilities	uint16;
	ServerLanguage		uint8;
	ServerStatus		uint16;

	Connected	bool;

	scrambleBuffer	[]byte;

	reader		*bufio.Reader;
	writer		*bufio.Writer;
	connection	net.Conn;

	database	string;
	username	string;
	password	string;
}


//Read initial handshake packet.
func (mysql *MySQLInstance) readInit() os.Error {
	var ph PacketHeaderR;
	binary.Read(mysql.reader, binary.LittleEndian, &ph);	//Header (Length and Seqence number)
	if ph.Seq != 0 {
		// Initial packet must be Seq == 0
		return os.ErrorString("Unexpected Sequence Number")
	}
	binary.Read(mysql.reader, binary.LittleEndian, &mysql.ProtocolVersion);
	mysql.ServerVersion, _ = mysql.reader.ReadString('\x00');
	binary.Read(mysql.reader, binary.LittleEndian, &mysql.ThreadId);
	var sb [9]byte;
	mysql.reader.Read(&sb);
	binary.Read(mysql.reader, binary.LittleEndian, &mysql.ServerCapabilities);
	binary.Read(mysql.reader, binary.LittleEndian, &mysql.ServerLanguage);
	binary.Read(mysql.reader, binary.LittleEndian, &mysql.ServerStatus);
	var sb2 [26]byte;
	mysql.reader.Read(&sb2);
	mysql.scrambleBuffer = new([20]byte);
	bytes.Copy(mysql.scrambleBuffer[0:8], sb[0:8]);
	bytes.Copy(mysql.scrambleBuffer[8:20], sb2[13:25]);
	return nil;
}

//Tries to read OK result error on error packett
func (mysql *MySQLInstance) readResult() os.Error {
	var ph PacketHeaderR;
	err := binary.Read(mysql.reader, binary.LittleEndian, &ph);
	var result byte;
	err = binary.Read(mysql.reader, binary.LittleEndian, &result);
	if result == 0xff {
		var errcode uint16;
		binary.Read(mysql.reader, binary.LittleEndian, &errcode);
		status := make([]byte, 6);
		mysql.reader.Read(status);
		msg, _ := mysql.reader.ReadString(0x00);
		return os.ErrorString(fmt.Sprintf("MySQL Error: (Code: %d) (Status: %s) %s", errcode, string(status), msg));
	}
	if err != nil {
		return err
	}
	fmt.Printf("Result == %x\n", result);
	return nil;
}

func (mysql *MySQLInstance) command(command MySQLCommand, arg string) os.Error	{
	plen := len(arg) + 1;
        var head [5]byte;
        head[0] = byte(plen);
        head[1] = byte(plen >> 8);
        head[2] = byte(plen >> 16);
        head[3] = 0;
	head[4] = uint8(command);
	_, err := mysql.writer.Write(&head);
	err = mysql.writer.WriteString(arg);
	err = mysql.writer.Flush();
	return err;
}
// Try to auth using the MySQL secure auth *crossing fingers*
func (mysql *MySQLInstance) sendAuth() os.Error {
	var clientFlags ClientFlags = CLIENT_LONG_PASSWORD + CLIENT_PROTOCOL_41 + CLIENT_SECURE_CONNECTION;
	var plen int = len(mysql.username);
	if len(mysql.database) > 0 {
		clientFlags += CLIENT_CONNECT_WITH_DB;
		plen += len(mysql.database) + 55;
	} else {
		plen += 54
	}
	var head [13]byte;
	head[0] = byte(plen);
	head[1] = byte(plen >> 8);
	head[2] = byte(plen >> 16);
	head[3] = 1;
	binary.LittleEndian.PutUint32(head[4:8], uint32(clientFlags));
	binary.LittleEndian.PutUint32(head[8:12], uint32(1073741824));
	head[12] = mysql.ServerLanguage;
	mysql.writer.Write(&head);
	var filler [23]byte;
	mysql.writer.Write(&filler);
	mysql.writer.WriteString(mysql.username);
	mysql.writer.Write(filler[0:1]);
	token := mysqlPassword(strings.Bytes(mysql.password), mysql.scrambleBuffer);
	mysql.writer.Write(token);
	if len(mysql.database) > 0 {
		mysql.writer.WriteString(mysql.database);
		mysql.writer.Write(filler[0:1]);
	}
	mysql.writer.Flush();

	return nil;

}

func (mysql *MySQLInstance) Query(arg string) os.Error {
	err := mysql.command(COM_QUERY, arg);
	return err;
}

//Connects to mysql server and reads the initial handshake,
//then tries to login using supplied credentials.
func Connect(host string, username string, password string, database string) (*MySQLInstance, os.Error) {
	var err os.Error;
	mysql := new(MySQLInstance);
	mysql.username = username;
	mysql.password = password;
	mysql.database = database;
	mysql.connection, err = net.Dial("tcp", "", host);
	if err != nil {
		return nil, os.ErrorString(fmt.Sprintf("Cant connect to %s\n", host))
	}
	mysql.reader = bufio.NewReader(mysql.connection);
	mysql.writer = bufio.NewWriter(mysql.connection);
	if err = mysql.readInit(); err != nil {
		return nil, err
	}
	err = mysql.sendAuth();
	if err = mysql.readResult(); err != nil {
		return nil, err
	}
	mysql.Connected = true;
	fmt.Printf("Connected to server\n");
	return mysql, nil;
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

