
TARGET=ons-client
SOURCES=main.c src/ons.c
LIBS=-lpthread -lczmq

${TARGET}: ${SOURCES}
	gcc ${SOURCES} ${LIBS} -o $@

run:
	./${TARGET} "inproc://ons" "testClient"

clean:
	rm ${TARGET}
