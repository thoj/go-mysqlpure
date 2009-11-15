GOFILES = mysql.go mysql_const.go mysql_util.go

all: $(GOARCH)

clean: 
	rm *.8 test

386:	
	8g $(GOFILES)
	8g test.go
	8l -o test test.8

x64:
	6g $(GOFILES)
	6g test.go
	6l -o test test.8

arm:
	5g $(GOFILES)
	5g test.go
	5l -o test test.8
test:
	./test


