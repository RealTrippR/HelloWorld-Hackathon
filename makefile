build/server.exe: main.go server/server.go api/api.go
	go build -o build/server.exe

run: build/server.exe
	./build/server.exe

clean:
	rm ./build/*.exe