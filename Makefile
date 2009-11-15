all: $(GOARCH)

clean: 
	rm *.8 test

386:	
	8g mysql.go 
	8g test.go
	8l -o test test.8

x64:
	6g mysql.go 
	6g test.go
	6l -o test test.8

arm:
	5g mysql.go 
	5g test.go
	5l -o test test.8


