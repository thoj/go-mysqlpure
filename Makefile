include $(GOROOT)/src/Make.inc

TARG=mysql
GOFILES=mysql.go mysql_const.go mysql_data.go mysql_util.go mysql_stmt.go

include $(GOROOT)/src/Make.pkg 
