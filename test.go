package main


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

type PacketHeaderR struct {
	Len1	uint8;
	Len2	uint8;
	Len3	uint8;
	Seq	uint8;
}


var ProtocolVersion uint8
var ServerVersion string
var ThreadId uint32
var ServerCapabilities uint16
var ServerLanguage uint8
var ServerStatus uint16


var scrambleBuffer []byte

//Read initial handshake packet.
func readInitPacket(br *bufio.Reader) os.Error {
	_ = binary.Read(br, binary.LittleEndian, &ProtocolVersion);
	ServerVersion, _ = br.ReadString('\x00');
	_ = binary.Read(br, binary.LittleEndian, &ThreadId);
	var sb [9]byte;
	_, _ = br.Read(&sb);
	_ = binary.Read(br, binary.LittleEndian, &ServerCapabilities);
	_ = binary.Read(br, binary.LittleEndian, &ServerLanguage);
	_ = binary.Read(br, binary.LittleEndian, &ServerStatus);
	var sb2 [26]byte;
	_, _ = br.Read(&sb2);
	scrambleBuffer = new([20]byte);
	_ = bytes.Copy(scrambleBuffer[0:8], sb[0:8]);
	_ = bytes.Copy(scrambleBuffer[8:20], sb2[13:25]);
	return nil;
}

//Tries to read incoming packet
func readPacket(br *bufio.Reader) os.Error {
	var ph PacketHeaderR;

	e := binary.Read(br, binary.LittleEndian, &ph);
	if ph.Seq == 0 {
		readInitPacket(br)
	}
	if e != nil {
		fmt.Printf("%s\n", e);
		os.Exit(1);
	}
	return nil;
}

//Connects to mysql server and reads the initial handshake, 
//then tries to login using supplied credentials.
func Connect(host string, database string, username string, password string) os.Error {
	conn, err := net.Dial("tcp", "", host);
	if err != nil {
		return os.ErrorString(fmt.Sprintf("Cant connect to %s\n", host))
	}
	br := bufio.NewReader(conn);
	bw := bufio.NewWriter(conn);
	if err = readPacket(br); err != nil {
		return err
	}
	sendAuth(bw, database, username, password);
	if err = readPacket(br); err != nil {
		return err
	}
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
		token[i+1] = result[i] ^ stage1[i];
	}
	for i := 0; i < 20; i++ {
		token_t[i] = token[i+1] ^ result[i]
	}
	token[0] = 20;
	return token;
}

func sendAuth(bw *bufio.Writer, database string, username string, password string) os.Error {
	var clientattr uint32 = CLIENT_LONG_PASSWORD + CLIENT_PROTOCOL_41 + CLIENT_SECURE_CONNECTION;
	var len int = len(username) + len(database) + 55;
	var head [13]byte;
	head[0] = byte(len);
	head[1] = byte(len >> 8);
	head[2] = byte(len >> 16);
	head[3] = 1;
	binary.LittleEndian.PutUint32(head[4:8], clientattr);
	binary.LittleEndian.PutUint32(head[8:12], uint32(1073741824));
	head[12] = ServerLanguage;
	_, _ = bw.Write(&head);
	fmt.Printf("%v\n", head);
	var filler [23]byte;
	_, _ = bw.Write(&filler);
	_ = bw.WriteString(username);
	_, _ = bw.Write(filler[0:1]);
	token := mysqlPassword(strings.Bytes(password));
	_, _ = bw.Write(token);
	_ = bw.WriteString(database);
	_, _ = bw.Write(filler[0:1]);
	bw.Flush();

	return nil;

}

func main() {
	_ = Connect("127.0.0.1:3306", "pre", "root", "omega51");
	fmt.Printf("Protocol = %d, Version = %s, Thread = %d, Capabilities = %d, Language = %d, Status = %d\n", ProtocolVersion, ServerVersion, ThreadId, ServerCapabilities, ServerLanguage, ServerStatus);
}
	
