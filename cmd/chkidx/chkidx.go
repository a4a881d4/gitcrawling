package main

import (
	"encoding/binary"
	"fmt"
	"os"
)

func main() {
	var err error
	var fr *os.File
	fr, err = os.Open(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	defer fr.Close()

	var mo = make(map[[32]byte]uint32)

	var head [32]byte
	var offset uint32
	for err == nil {
		var n int
		n, err = fr.Read(head[:])
		if err != nil {
			fmt.Println("Read", err)
			break
		}
		if n != 32 {
			fmt.Println("Read", n)
			return
		}
		if _, ok := mo[head]; ok {
			fmt.Println("Dup @", offset)
		}
		mo[head] = offset
		size := binary.BigEndian.Uint32(head[20:24])
		// fmt.Println("Size @", size, offset)
		offset += size + 32
		_, err = fr.Seek(int64(size), 1)
		if err != nil {
			fmt.Println("Seek", err)
			break
		}
	}

	// idxf := strings.Replace(os.Args[1], ".pack", ".idx", -1)

}
