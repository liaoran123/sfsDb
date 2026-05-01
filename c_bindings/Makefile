CC = gcc
CFLAGS = -Wall -Wextra -I.
LIBS = -L. -lsfsdb -lmingwex -lmingw32

all: test_libsfsdb

test_libsfsdb: test_libsfsdb.c libsfsdb.a
	$(CC) $(CFLAGS) -o $@ $< $(LIBS)

clean:
	rm -f test_libsfsdb.exe
