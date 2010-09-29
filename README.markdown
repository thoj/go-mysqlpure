Project Goal
----------------
The goal of this project is to implement the MySQL wire protocol in Go,
mostly for my own amusement but it might become usable as a client 
library for other Go projects.

The wire protocol is documented [here](http://forge.mysql.com/wiki/MySQL_Internals_ClientServer_Protocol "MySQL Wire Protocol Documentation")

Status
---------------
* Most queries work
* Server side prepared statements work. (With common types)
* See example/simple.go for simple example
* See example/prepared.go for example using server side prepared statements

Install
--------------
	$ git clone git://github.com/thoj/Go-MySQL-Client-Library.git
	$ cd Go-MySQL-Client-Library
	$ make
	$ make install

Use
--------------
Three first parameters are passed to Dial. Unix socket: net = unix, raddr = path to mysql.sock  
> dbh, error = mysql.Connect(net, laddr, raddr, username, password, database)

Select database  
> res, err = dbh.Use(database)

Run simple Query. AffectedRows and InsertId is in res  
> res, err = dbh.Query(sql)

Prepare server side statement  
> sth, err = dbh.Prepare(<SQL with ? placeholders>)

Execute prepared statement (Only supports string, int, float parameters):  
> res, err = sth.Execute(parameters ...)

Fetch row from query with result set  
> row, err = res.FetchRow()

Fetch one row as map[string]string  
> rowmap = res.FetchRowMap()

Fetch all rows as []map[string]string  
> rowsmap = res.FetchAllRowMap()

FAQ
----------
* Q: I'm getting question marks instead of my Unicode characters
* A: Run `dbh.Query("SET NAMES utf8")` before the select query 
