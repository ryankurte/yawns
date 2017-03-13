
TARGET=ons-client
LIB=libons
SOURCES=src/ons.c
OUTDIR=build

CC=gcc
AR=ar
CFLAGS=-c -Wall -pedantic
LFLAGS=-lpthread -lczmq -lzmq

# Map source files to objects
OBJECTS=$(SOURCES:src/%.c=$(OUTDIR)/%.o)

# Helper to set default action
all: libs util

# Helper to create output dir
dir:
	@mkdir -p $(OUTDIR)

# Generic compilation rule
$(OUTDIR)/%.o: src/%.c dir
	@$(CC) $(INCLUDES) $(CFLAGS) $< -o $@

# Library helper
libs: $(OUTDIR)/$(LIB).a $(OUTDIR)/$(LIB).so

# Create static library
$(OUTDIR)/$(LIB).a: $(OBJECTS)
	@$(AR) rcs $@ $^

# Create shared library
$(OUTDIR)/$(LIB).so: $(OBJECTS)
	@$(CC) -shared $(LFLAGS) $^ -o $@ 

# Util helper
util: $(OUTDIR)/$(TARGET)

# Build and link example util
$(OUTDIR)/$(TARGET): $(OUTDIR)/main.o $(LIB).a
	@$(CC) $^ $(LFLAGS) -o $@

# Run the example util
run:
	./$(TARGET) "ipc:///tmp/ons" "testClient"

# Cleanup build files
clean:
	rm -rf $(OUTDIR)
