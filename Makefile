main: main.go
	go build

run: main
	time ./main
