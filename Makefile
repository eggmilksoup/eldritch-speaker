all:
	mkdir -p bin
	${MAKE} bin/drafting bin/speaker

bin/drafting: src/drafting.go
	go build -o bin/drafting src/drafting.go

bin/dummy:
	ln -sf /bin/true bin/dummy

bin/speaker: src/speaker.go
	go build -o bin/speaker src/speaker.go

bin/rcp: src/rcp.go
	go build -o bin/rcp src/rcp.go
