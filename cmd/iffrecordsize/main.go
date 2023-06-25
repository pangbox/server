// iffrecordsize implements a small algorithm that can determine the size of a
// record in a non-empty IFF file, by dividing its filesize sans header with the
// number of records in the header. It accepts multiple files at once, making it
// easy to compare against a lot of different versions of files at once.
package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"os"
)

func main() {
	flag.Parse()
	for _, name := range flag.Args() {
		load(name)
	}
}

func load(name string) {
	finfo, err := os.Stat(name)
	if err != nil {
		fmt.Printf("%s: stat: %v\n", name, err)
		return
	}
	filesize := finfo.Size()
	file, err := os.Open(name)
	if err != nil {
		fmt.Printf("%s: open: %v\n", name, err)
		return
	}
	var header [8]byte
	_, err = file.Read(header[:])
	file.Close()
	if err != nil {
		fmt.Printf("%s: read: %v\n", name, err)
		return
	}
	recordcount := int64(binary.LittleEndian.Uint16(header[0:2]))
	if recordcount == 0 {
		fmt.Printf("%s: no records\n", name)
		return
	}
	bodysize := filesize - int64(len(header))
	if bodysize%recordcount != 0 {
		fmt.Printf("%s: size not divisible by record count: %d\n", name, bodysize%recordcount)
		return
	}
	recordsize := bodysize / recordcount
	fmt.Printf("%s: %d (%08x)\n", name, recordsize, recordsize)
}
