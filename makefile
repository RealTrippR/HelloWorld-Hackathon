build/server.exe:
	go build -o server.exe

run: build/server.exe
	./build/server.exe

clean:
	rm ./build/*.exe