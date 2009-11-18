Project Goal
----------------
The goal of this project is to implemnt the MySQL wire protocol in Go,
mostly for my own amusement but it might become usable as a client 
library for other Go projects.

The wire protocol is documented here: 
 http://forge.mysql.com/wiki/MySQL_Internals_ClientServer_Protocol

Status
---------------
Very simple queries should work. See test.go for example. 

Use
--------------
dbh, error = mysql.Connect(net, laddr, raddr, username, password, database)
Three first parameters are passed to Dial. Unix socket: net = unix, raddr = path to mysql.sock

dbh.Use(database)
Select database

res, err = dbh.Query(sql)
Run Query. AffectedRows and InsertId is in res

row = res.FetchRow();
Fetch row from query with resultset.


### Output of test:
Connected to 5.0.77-1
Response = OK, Affected Rows = 1, Insert Id = 11, Server Status = 2
Response = OK, Affected Rows = 1, Server Status = 2
Response = ResultSet, Server Status = 0
id      sha1
1       a1b7edd205324a89803ed96de72f072201797ce0

