
TARGET=ons-client
LIB=libons
SOURCES=src/ons.c
OUTDIR=build

CC=gcc
AR=ar
CFLAGS=-c -Wall -pedantic
LFLAGS=-lpthread -lczmq -lzmq
IFLAGS=-I./

# Map source files to objects
OBJECTS=$(SOURCES:src/%.c=$(OUTDIR)/%.o)

# Helper to set default action
all: libs client

# Helper to create output dir
dir:
	@mkdir -p $(OUTDIR)

# Generic compilation rule
$(OUTDIR)/%.o: src/%.c dir
	@$(CC) $(IFLAGS) $(CFLAGS) $< -o $@

# Library helper
libs: $(OUTDIR)/$(LIB).a $(OUTDIR)/$(LIB).so

# Create static library
$(OUTDIR)/$(LIB).a: $(OBJECTS)
	@$(AR) rcs $@ $^

# Create shared library
$(OUTDIR)/$(LIB).so: $(OBJECTS)
	@$(CC) -shared $(LFLAGS) $^ -o $@ 

# Client helper
client: $(OUTDIR)/$(TARGET)

# Build and link example client
$(OUTDIR)/$(TARGET): $(OUTDIR)/main.o $(OUTDIR)/$(LIB).a
	@$(CC) $^ $(LFLAGS) -o $@

# Run the example client
run:
	./$(TARGET) "ipc:///tmp/ons" "testClient"

# Cleanup build files
clean:
	rm -rf $(OUTDIR)
