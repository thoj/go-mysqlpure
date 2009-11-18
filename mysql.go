// Copyright 2009 Thomas Jager <mail@jager.no>  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// MySQL 4.1+ Client Library.

package mysql

import (
	"net";
	"os";
	"bytes";
	"bufio";
	"encoding/binary";
	"strings";
	"fmt";
)


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
	ph := readHeader(mysql.reader);

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


func (res *MySQLResponse) readRowPacket(br *bufio.Reader) *MySQLRow {
	readHeader(br);
	row := new(MySQLRow);
	row.Data = make([]*MySQLData, res.ResultSet.FieldCount);
	if peekEOF(br) { //FIXME: Ignoring EOF and return nil is a bit hackish.
		ignore := make([]byte, 5);
		br.Read(ignore);
		return nil
	}
	for i := uint64(0); i < res.ResultSet.FieldCount; i++ {
		s, isnull := unpackString(br);
		data := new(MySQLData);
		data.IsNull = isnull;
		data.Data = s;
		data.Length = uint64(len(s));
		data.Type = res.ResultSet.Fields[i].Type;
		row.Data[i] = data;
	}
	return row;
}

func (mysql *MySQLInstance) readResultSet(fieldCount uint64) (*MySQLResultSet, os.Error) {
	rs := new(MySQLResultSet);
	rs.FieldCount = fieldCount;
	rs.Fields = make([]*MySQLField, rs.FieldCount);
	var i uint64;
	for i = 0; i < rs.FieldCount; i++ {
		readHeader(mysql.reader);
		rs.Fields[i] = readFieldPacket(mysql.reader);
	}
	readEOFPacket(mysql.reader);
	return rs, nil;
}

//Tries to read OK result error on error packett
func (mysql *MySQLInstance) readResult() (*MySQLResponse, os.Error) {
	ph := readHeader(mysql.reader);
	if ph.Len < 1 {
		return nil, os.ErrorString("Packet to small")
	}
	response := new(MySQLResponse);
	response.EOF = false;
	err := binary.Read(mysql.reader, binary.LittleEndian, &response.FieldCount);
	if response.FieldCount == 0xff {	// ERROR
		var errcode uint16;
		binary.Read(mysql.reader, binary.LittleEndian, &errcode);
		status := make([]byte, 6);
		mysql.reader.Read(status);
		msg := make([]byte, ph.Len-1-2-6);
		mysql.reader.Read(msg);
		return nil, os.ErrorString(fmt.Sprintf("MySQL Error: (Code: %d) (Status: %s) %s", errcode, string(status), string(msg)));

	} else if response.FieldCount == 0x00 {	// OK
		eb, _ := unpackLength(mysql.reader);
		response.AffectedRows = eb;
		eb, _ = unpackLength(mysql.reader);
		response.InsertId = eb;
		err = binary.Read(mysql.reader, binary.LittleEndian, &response.ServerStatus);
		err = binary.Read(mysql.reader, binary.LittleEndian, &response.WarningCount);

	} else if response.FieldCount > 0x00 && response.FieldCount < 0xFB {	//Result|Field|Row Data
		rs, _ := mysql.readResultSet(uint64(response.FieldCount));
		response.ResultSet = rs;
		return response, err;

	} else if response.FieldCount == 0xFE {	// EOF
		err = binary.Read(mysql.reader, binary.LittleEndian, &response.ServerStatus);
		err = binary.Read(mysql.reader, binary.LittleEndian, &response.WarningCount);
		response.EOF = true;
		return response, err;
	}
	if err != nil {
		return nil, err
	}
	return response, nil;
}

func (mysql *MySQLInstance) command(command MySQLCommand, arg string) (*MySQLResponse, os.Error) {
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
	if err != nil {
		return nil, err
	}
	if command == COM_QUIT { // Don't bother reading anything more.
		return nil, nil;
	}

	return mysql.readResult();
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
	if len(mysql.password) < 1 {
		plen -= 20
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
	if len(mysql.password) > 0 {
		token := mysqlPassword(strings.Bytes(mysql.password), mysql.scrambleBuffer);
		mysql.writer.Write(token);
	} else {
		mysql.writer.Write(filler[0:1])
	}
	if len(mysql.database) > 0 {
		mysql.writer.WriteString(mysql.database);
		mysql.writer.Write(filler[0:1]);
	}
	mysql.writer.Flush();

	return nil;

}
func (mysql *MySQLInstance) Use(arg string)	{ mysql.command(COM_INIT_DB, arg) }
func (mysql *MySQLInstance) Quit()		{ mysql.command(COM_QUIT, "") }

func (rs *MySQLResponse) FetchRow() *MySQLRow	{ return rs.readRowPacket(rs.mysql.reader) }

func (mysql *MySQLInstance) Query(arg string) (*MySQLResponse, os.Error) {
	response := new(MySQLResponse);
	response, err := mysql.command(COM_QUERY, arg);
	if response != nil {
		response.mysql = mysql
	}
	return response, err;
}

//Connects to mysql server and reads the initial handshake,
//then tries to login using supplied credentials.
//The first 3 parameters are passed directly to Dial
func Connect(netstr string, laddrstr string, raddrstr string, username string, password string, database string) (*MySQLInstance, os.Error) {
	var err os.Error;
	mysql := new(MySQLInstance);
	mysql.username = username;
	mysql.password = password;
	mysql.database = database;
	mysql.connection, err = net.Dial(netstr, laddrstr, raddrstr);
	if err != nil {
		return nil, os.ErrorString(fmt.Sprintf("Cant connect to %s\n", raddrstr))
	}
	mysql.reader = bufio.NewReader(mysql.connection);
	mysql.writer = bufio.NewWriter(mysql.connection);
	if err = mysql.readInit(); err != nil {
		return nil, err
	}
	err = mysql.sendAuth();
	if _, err = mysql.readResult(); err != nil {
		return nil, err
	}
	mysql.Connected = true;
	return mysql, nil;
}
