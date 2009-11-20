Project Goal
----------------
The goal of this project is to implemnt the MySQL wire protocol in Go,
mostly for my own amusement but it might become usable as a client 
library for other Go projects.

The wire protocol is documented here: 
 http://forge.mysql.com/wiki/MySQL_Internals_ClientServer_Protocol

Status
---------------
* Most queries should work. 
* Password auth works.

See example/test.go for example.

Install
--------------
	$ git clone git@github.com:thoj/Go-MySQL-Client-Library.git
	$ cd Go-MySQL-Client-Library
	$ make
	$ make install

Use
--------------
Three first parameters are passed to Dial. Unix socket: net = unix, raddr = path to mysql.sock
	dbh, error = mysql.Connect(net, laddr, raddr, username, password, database)

Select database
	dbh.Use(database)

Run Query. AffectedRows and InsertId is in res
	res, err = dbh.Query(sql)

Fetch row from query with resultset.
	row = res.FetchRow();

Fetch Row map[string]string from query with resultset.
	rowmap = res.FetchRow();

