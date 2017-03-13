
TARGET=ons-client
SOURCES=main.c src/ons.c
LIBS=-lpthread -lczmq -lzmq

${TARGET}: ${SOURCES}
	gcc ${SOURCES} ${LIBS} -o $@

run:
	./${TARGET} "ipc:///tmp/ons" "testClient"

clean:
	rm ${TARGET}
