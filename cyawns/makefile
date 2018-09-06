# C-ONS Example / Library Makefile

TARGET=yawns-client
LIB=libyawns
OUTDIR=build

# Helper to set default action
all: dir
	cd $(OUTDIR) && cmake .. && make && cd ..

# Helper to create output dir
dir:
	@mkdir -p $(OUTDIR)/protocol

# Run the example client
run:
	./$(TARGET) "ipc:///tmp/ons" "testClient"

test: all
	GODEBUG=cgocheck=0 go test -v -p 1 -timeout 10s ./

install:
	cd $(OUTDIR) && cmake .. && make install; cd ..

# Cleanup build files
clean:
	rm -rf $(OUTDIR)


