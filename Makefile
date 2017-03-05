main: main.go
	go build main.go

run: main
	time ./main
