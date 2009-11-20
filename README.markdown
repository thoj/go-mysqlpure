Project Goal
----------------
The goal of this project is to implemnt the MySQL wire protocol in Go,
mostly for my own amusement but it might become usable as a client 
library for other Go projects.

The wire protocol is documented here: 
 http://forge.mysql.com/wiki/MySQL_Internals_ClientServer_Protocol

Status
---------------
Simple queries should work. See test.go for example.
Password auth works.

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


### Output of test:
	Connected to 5.0.77-1
	Response = OK, Affected Rows = 1, Insert Id = 11, Server Status = 2
	Response = OK, Affected Rows = 1, Server Status = 2
	Response = ResultSet, Server Status = 0
	id      sha1
	1       a1b7edd205324a89803ed96de72f072201797ce0

