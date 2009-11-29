// Copyright 2009 Thomas Jager <mail@jager.no>  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mysql

const (
	MAX_PACKET_SIZE = (1 << 24);
)

type ClientFlags uint32

const (
	CLIENT_LONG_PASSWORD		ClientFlags	= 1;		/* new more secure passwords */
	CLIENT_FOUND_ROWS				= 2;		/* Found instead of affected rows */
	CLIENT_LONG_FLAG				= 4;		/* Get all column flags */
	CLIENT_CONNECT_WITH_DB				= 8;		/* One can specify db on connect */
	CLIENT_NO_SCHEMA				= 16;		/* Don't allow database.table.column */
	CLIENT_COMPRESS					= 32;		/* Can use compression protocol */
	CLIENT_ODBC					= 64;		/* Odbc client */
	CLIENT_LOCAL_FILES				= 128;		/* Can use LOAD DATA LOCAL */
	CLIENT_IGNORE_SPACE				= 256;		/* Ignore spaces before '(' */
	CLIENT_PROTOCOL_41				= 512;		/* New 4.1 protocol */
	CLIENT_INTERACTIVE				= 1024;		/* This is an interactive client */
	CLIENT_SSL					= 2048;		/* Switch to SSL after handshake */
	CLIENT_IGNORE_SIGPIPE				= 4096;		/* IGNORE sigpipes */
	CLIENT_TRANSACTIONS				= 8192;		/* Client knows about transactions */
	CLIENT_RESERVED					= 16384;	/* Old flag for 4.1 protocol  */
	CLIENT_SECURE_CONNECTION			= 32768;	/* New 4.1 authentication */
	CLIENT_MULTI_STATEMENTS				= 65536;	/* Enable/disable multi-stmt support */
	CLIENT_MULTI_RESULTS				= 131072;	/* Enable/disable multi-results */
)


const (
	MYSQL_TYPE_DECIMAL	uint8	= iota;
	MYSQL_TYPE_TINY;
	MYSQL_TYPE_SHORT;
	MYSQL_TYPE_LONG;
	MYSQL_TYPE_FLOAT;
	MYSQL_TYPE_DOUBLE;
	MYSQL_TYPE_NULL;
	MYSQL_TYPE_TIMESTAMP;
	MYSQL_TYPE_LONGLONG;
	MYSQL_TYPE_INT24;
	MYSQL_TYPE_DATE;
	MYSQL_TYPE_TIME;
	MYSQL_TYPE_DATETIME;
	MYSQL_TYPE_YEAR;
	MYSQL_TYPE_NEWDATE;
	MYSQL_TYPE_VARCHAR;
	MYSQL_TYPE_BIT;
	MYSQL_TYPE_NEWDECIMAL	= 246;
	MYSQL_TYPE_ENUM		= 247;
	MYSQL_TYPE_SET		= 248;
	MYSQL_TYPE_TINY_BLOB	= 249;
	MYSQL_TYPE_MEDIUM_BLOB	= 250;
	MYSQL_TYPE_LONG_BLOB	= 251;
	MYSQL_TYPE_BLOB		= 252;
	MYSQL_TYPE_VAR_STRING	= 253;
	MYSQL_TYPE_STRING	= 254;
	MYSQL_TYPE_GEOMETRY	= 255;
)


type MySQLCommand uint32

const (
	COM_SLEEP			MySQLCommand	= iota;	//(none, this is an internal thread state)
	COM_QUIT;					//mysql_close
	COM_INIT_DB;					//mysql_select_db
	COM_QUERY;					//mysql_real_query
	COM_FIELD_LIST;					//mysql_list_fields
	COM_CREATE_DB;					//mysql_create_db (deprecated)
	COM_DROP_DB;					//mysql_drop_db (deprecated)
	COM_REFRESH;					//mysql_refresh
	COM_SHUTDOWN;					//mysql_shutdown
	COM_STATISTICS;					//mysql_stat
	COM_PROCESS_INFO;				//mysql_list_processes
	COM_CONNECT;					//(none, this is an internal thread state)
	COM_PROCESS_KILL;				//mysql_kill
	COM_DEBUG;					//mysql_dump_debug_info
	COM_PING;					//mysql_ping
	COM_TIME;					//(none, this is an internal thread state)
	COM_DELAYED_INSERT;				//(none, this is an internal thread state)
	COM_CHANGE_USER;				//mysql_change_user
	COM_BINLOG_DUMP;				//sent by the slave IO thread to request a binlog
	COM_TABLE_DUMP;					//LOAD TABLE ... FROM MASTER (deprecated)
	COM_CONNECT_OUT;				//(none, this is an internal thread state)
	COM_REGISTER_SLAVE;				//sent by the slave to register with the master (optional)
	COM_STMT_PREPARE;				//mysql_stmt_prepare
	COM_STMT_EXECUTE;				//mysql_stmt_execute
	COM_STMT_SEND_LONG_DATA;			//mysql_stmt_send_long_data
	COM_STMT_CLOSE;					//mysql_stmt_close
	COM_STMT_RESET;					//mysql_stmt_reset
	COM_SET_OPTION;					//mysql_set_server_option
	COM_STMT_FETCH;					//mysql_stmt_fetch
)
