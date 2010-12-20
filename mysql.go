// Copyright 2009 Thomas Jager <mail@jager.no>  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// MySQL 4.1+ Client Library.

package mysql

import (
	"net"
	"os"
	"bufio"
	"encoding/binary"
	"fmt"
        "sync"
)


type MySQLInstance struct {
	ProtocolVersion    uint8  // Protocol version = 0x10
	ServerVersion      string // Server string
	ThreadId           uint32 // Current Thread ID
	ServerCapabilities uint16
	ServerLanguage     uint8
	ServerStatus       uint16

	Connected    bool
	queryDone    bool
	lastResponse *MySQLResponse

	scrambleBuffer []byte

	reader     *bufio.Reader
	writer     *bufio.Writer
	connection net.Conn

	database string
	username string
	password string

        mutex *sync.Mutex
}


//Read initial handshake packet.
func (mysql *MySQLInstance) readInit() os.Error {
	ph, err := readHeader(mysql.reader)
	if err != nil {
		return err
	}

	if ph.Seq != 0 {
		// Initial packet must be Seq == 0
		return os.ErrorString("Unexpected Sequence Number")
	}
	err = binary.Read(mysql.reader, binary.LittleEndian, &mysql.ProtocolVersion)
	if err != nil { return err }
	mysql.ServerVersion, err = mysql.reader.ReadString('\x00')
	if err != nil { return err }
	err = binary.Read(mysql.reader, binary.LittleEndian, &mysql.ThreadId)
	if err != nil { return err }
	mysql.scrambleBuffer = make([]byte, 20)
	err = readFull(mysql.reader, mysql.scrambleBuffer[0:8])
        if err != nil { return err }
	err = ignoreBytes(mysql.reader, 1)
        if err != nil { return err }
	err = binary.Read(mysql.reader, binary.LittleEndian, &mysql.ServerCapabilities)
	if err != nil { return err }
	err = binary.Read(mysql.reader, binary.LittleEndian, &mysql.ServerLanguage)
	if err != nil { return err }
	err = binary.Read(mysql.reader, binary.LittleEndian, &mysql.ServerStatus)
	if err != nil { return err }
	err = ignoreBytes(mysql.reader, 13)
	if err != nil { return err }
	err = readFull(mysql.reader, mysql.scrambleBuffer[8:20])
        if err != nil { return err }
	err = ignoreBytes(mysql.reader, 1)
	return err
}


func (res *MySQLResponse) readRowPacket(br *bufio.Reader) (row *MySQLRow, err os.Error) {
	var (
		ph *PacketHeader
		eof bool
	)
	ph, err = readHeader(br)
	if err != nil { return }
	row = new(MySQLRow)
	eof, err = peekEOF(br)
	if err != nil { return }
	if eof || res.ResultSet == nil { //FIXME: Ignoring EOF and return nil is a bit hackish.
		err = ignoreBytes(br, ph.Len)
		return nil, err
	}
	row.Data = make([]*MySQLData, res.ResultSet.FieldCount)
	if res.Prepared {
		//TODO: Do this right.
		err = ignoreBytes(br, uint64(res.ResultSet.FieldCount+9)/8+1)
		if err != nil { return }
	}
	for i := uint64(0); i < res.ResultSet.FieldCount; i++ {
		data := new(MySQLData)
		var s string
		var isnull bool
		if res.Prepared {
			s, isnull, err = readFieldData(br, res.ResultSet.Fields[i])
		} else {
			s, isnull, err = unpackString(br)
		}
		if err != nil { return }
		data.IsNull = isnull
		data.Data = s
		data.Length = uint64(len(s))
		data.Type = res.ResultSet.Fields[i].Type
		row.Data[i] = data
	}
	return row, err
}

func (mysql *MySQLInstance) readResultSet(fieldCount uint64) (*MySQLResultSet, os.Error) {
	rs := new(MySQLResultSet)
	rs.FieldCount = fieldCount
	rs.Fields = make([]*MySQLField, rs.FieldCount)
	var (
		i uint64
		err os.Error
	)
	for i = 0; i < rs.FieldCount; i++ {
		_, err = readHeader(mysql.reader)
		if err != nil {
			return nil, err
		}
		rs.Fields[i], err = readFieldPacket(mysql.reader)
		if err != nil {
			return nil, err
		}
	}
	err = readEOFPacket(mysql.reader)
	return rs, err
}

//Tries to read OK result error on error packet
func (mysql *MySQLInstance) readResult() (*MySQLResponse, os.Error) {
	if mysql == nil {
		panic("mysql undefined")
	}
	ph, err := readHeader(mysql.reader)
	if err != nil {
		return nil, os.ErrorString(fmt.Sprintf("readHeader error: %s", err))
	} else if ph.Len < 1 {
		// Junk?
	}
	response := new(MySQLResponse)
	response.EOF = false
	response.FieldCount, _, err = unpackFieldCount(mysql.reader)
	if err != nil {
		return nil, err
	}
	response.mysql = mysql

	if response.FieldCount == 0xff { // ERROR
		return nil, readErrorPacket(mysql.reader, int(ph.Len))

	} else if response.FieldCount == 0x00 { // OK
		msg_len := int(ph.Len - 1)
		eb, _, num, err := unpackLength(mysql.reader)
		if err != nil {
			return nil, err
		}
		response.AffectedRows = eb
		msg_len -= num
		eb, _, num, err = unpackLength(mysql.reader)
		if err != nil {
			return nil, err
		}
		response.InsertId = eb
		msg_len -= num
		err = binary.Read(mysql.reader, binary.LittleEndian, &response.ServerStatus)
		if err != nil {
			return nil, err
		}
		err = binary.Read(mysql.reader, binary.LittleEndian, &response.WarningCount)
		if err != nil {
			return nil, err
		}
		msg_len -= 4
		mBuf := make([]byte, msg_len)
		err = readFull(mysql.reader, mBuf)
                if err != nil {
			return nil, err
		}
		response.Message = string(mBuf)
	} else if response.FieldCount > 0x00 && response.FieldCount < 0xFB { //Result|Field|Row Data
		rs, err := mysql.readResultSet(uint64(response.FieldCount))
		response.ResultSet = rs
		return response, err

	} else if response.FieldCount == 0xFE { // EOF
		err = binary.Read(mysql.reader, binary.LittleEndian, &response.ServerStatus)
		err = binary.Read(mysql.reader, binary.LittleEndian, &response.WarningCount)
		response.EOF = true
		return response, err

	}
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (dbh *MySQLInstance) mysqlCommand(command MySQLCommand, arg string) (*MySQLResponse, os.Error) {
	plen := len(arg) + 1
	head := make([]byte, 5)
	head[0] = byte(plen)
	head[1] = byte(plen >> 8)
	head[2] = byte(plen >> 16)
	head[3] = 0
	head[4] = uint8(command)
	_, err := dbh.writer.Write(head)
	_, err = dbh.writer.WriteString(arg)
	if err = dbh.writer.Flush(); err != nil {
		return nil, err
	}

	if command == COM_QUIT { // Don't bother reading anything more.
		return nil, nil
	}

	return dbh.readResult()
}


// Auth using the MySQL secure auth
func (dbh *MySQLInstance) sendAuth() os.Error {
	var clientFlags ClientFlags = CLIENT_LONG_PASSWORD
	clientFlags += CLIENT_PROTOCOL_41
	clientFlags += CLIENT_SECURE_CONNECTION
	clientFlags += CLIENT_MULTI_STATEMENTS
	clientFlags += CLIENT_MULTI_RESULTS
	var plen int = len(dbh.username)
	if len(dbh.database) > 0 {
		clientFlags += CLIENT_CONNECT_WITH_DB
		plen += len(dbh.database) + 55
	} else {
		plen += 54
	}
	if len(dbh.password) < 1 {
		plen -= 20
	}
	head := make([]byte, 13)
	head[0] = byte(plen)
	head[1] = byte(plen >> 8)
	head[2] = byte(plen >> 16)
	head[3] = 1
	binary.LittleEndian.PutUint32(head[4:8], uint32(clientFlags))
	binary.LittleEndian.PutUint32(head[8:12], uint32(MAX_PACKET_SIZE))
	head[12] = dbh.ServerLanguage
	_, err := dbh.writer.Write(head)
	if err != nil { return err }
	filler := make([]byte, 23)
	_, err = dbh.writer.Write(filler)
	if err != nil { return err }
	_, err = dbh.writer.WriteString(dbh.username)
	if err != nil { return err }
	_, err = dbh.writer.Write(filler[0:1])
	if err != nil { return err }
	if len(dbh.password) > 0 {
		token := mysqlPassword([]byte(dbh.password), dbh.scrambleBuffer)
		_, err = dbh.writer.Write(token)
	} else {
		_, err = dbh.writer.Write(filler[0:1])
	}
	if err != nil { return err }
	if len(dbh.database) > 0 {
		_, err = dbh.writer.WriteString(dbh.database)
		if err != nil { return err }
		_, err = dbh.writer.Write(filler[0:1])
		if err != nil { return err }
	}
	err = dbh.writer.Flush()

	return err

}
//Stolen from http://golang.org/doc/effective_go.html#slices
func appendMap(slice, data []map[string]string) []map[string]string {
	l := len(slice)
	if l+len(data) > cap(slice) { // reallocate
		// Allocate double what's needed, for future growth.
		newSlice := make([]map[string]string, (l+len(data))*PRE_ALLOCATE)
		// Copy data (could use bytes.Copy()).
		for i, c := range slice {
			newSlice[i] = c
		}
		slice = newSlice
	}
	slice = slice[0 : l+len(data)]
	for i, c := range data {
		slice[l+i] = c
	}
	return slice
}

//Connects to mysql server and reads the initial handshake,
//then tries to login using supplied credentials.
//The first 3 parameters are passed directly to Dial
func Connect(netstr string, laddrstr string, raddrstr string, username string, password string, database string) (*MySQLInstance, os.Error) {
	var err os.Error
	dbh := new(MySQLInstance)
        dbh.mutex = new(sync.Mutex)
	dbh.username = username
	dbh.password = password
	dbh.database = database
	dbh.connection, err = net.Dial(netstr, laddrstr, raddrstr)
	if err != nil {
		return nil, os.ErrorString(fmt.Sprintf("Cant connect to %s\n", raddrstr))
	}
	dbh.reader = bufio.NewReader(dbh.connection)
	dbh.writer = bufio.NewWriter(dbh.connection)
	if err = dbh.readInit(); err != nil {
		return nil, err
	}
	err = dbh.sendAuth()
	if _, err = dbh.readResult(); err != nil {
		return nil, err
	}
	dbh.Connected = true
	dbh.queryDone = true
	return dbh, nil
}

func (dbh *MySQLInstance) Use(arg string) (*MySQLResponse, os.Error) {
	if dbh == nil {
		panic("dbh object is undefined")
	}
	return dbh.mysqlCommand(COM_INIT_DB, arg)
}

func (dbh *MySQLInstance) Quit() {
	if dbh == nil {
		panic("dbh object is undefined")
	}
	dbh.mysqlCommand(COM_QUIT, "")
	dbh.connection.Close()
}

func (dbh *MySQLInstance) Prepare(arg string) (*MySQLStatement, os.Error) {
	if dbh == nil {
		panic("dbh object is undefined")
	}
	return dbh.prepare(arg)
}

const (
	PRE_ALLOCATE = 30
)

//Fetches all rows from result
func (rs *MySQLResponse) FetchAllRowMap() []map[string]string {
	rr := make([]map[string]string, PRE_ALLOCATE) // Good tradeoff? Probably not.
	tmp := make([]map[string]string, 1)           //What?
	row := 0
	for r := rs.FetchRowMap(); r != nil; r = rs.FetchRowMap() {
		if row < PRE_ALLOCATE {
			rr[row] = r
		} else {
			tmp[0] = r
			rr = appendMap(rr, tmp)
		}
		row++

	}
	return rr[0:row]
}

//Fetch next row.
func (rs *MySQLResponse) FetchRow() (*MySQLRow, os.Error) {
	row, err := rs.readRowPacket(rs.mysql.reader)
	if err != nil {
		rs.mysql.queryDone = true
		return nil, err
	} else if row == nil {
		rs.mysql.queryDone = true
		return nil, nil
	}
	rs.mysql.queryDone = false //Hack. Make sure this is false as long as there may be row packets in the buffer!
	return row, nil
}

//Fetch next row map.
func (rs *MySQLResponse) FetchRowMap() map[string]string {
	if rs == nil {
		panic("rs undefined")
	}
	row, err := rs.FetchRow()
	if row == nil || err != nil {
		rs.mysql.queryDone = true
		return nil
	}
	m := make(map[string]string)
	for i := 0; i < len(row.Data); i++ {
		m[rs.ResultSet.Fields[i].Name] = row.Data[i].Data
	}
	return m
}

//Send query to server and read response. Return response object.
func (dbh *MySQLInstance) Query(arg string) (*MySQLResponse, os.Error) {
	if dbh == nil {
		panic("dbh object is undefined")
	}
        // For multithreading
        dbh.mutex.Lock()
        defer dbh.mutex.Unlock()

	// Terrible hack to handle unfinished Queries.
	if dbh.queryDone == false {
		for r, _ := dbh.lastResponse.FetchRow(); r != nil; r, _ = dbh.lastResponse.FetchRow() {
		}
	}
	response := new(MySQLResponse)
	response, err := dbh.mysqlCommand(COM_QUERY, arg)
	if response != nil {
		response.mysql = dbh
		dbh.lastResponse = response
	}
	return response, err
}

func (sth *MySQLStatement) Execute(va ...interface{}) (*MySQLResponse, os.Error) {
	return sth.execute(va)
}
