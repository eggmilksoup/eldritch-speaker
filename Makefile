all:
	mkdir -p bin
	${MAKE} bin/drafting bin/speaker

bin/drafting: src/drafting.go
	go build -o bin/drafting src/drafting.go

bin/speaker: src/speaker.go
	go build -o bin/speaker src/speaker.go
