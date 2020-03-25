package main

// func getOffset(idx *idxfile.MemoryIndex, v []byte) uint64 {
// 	ofs := encbin.BigEndian.Uint32(v[:])

// 	if (uint64(ofs) & isO64Mask) != 0 {
// 		offset := 8 * (uint64(ofs) & ^isO64Mask)
// 		n := encbin.BigEndian.Uint64(idx.Offset64[offset : offset+8])
// 		return n
// 	}

// 	return uint64(ofs)
// }
/*
	for l := 0; l < 256; l++ {
		i := idx.FanoutMapping[l]
		if i == noMapping {
			continue
		}
		num := len(idx.Offset32[i]) >> 2
		for k := 0; k < num; k++ {
			offset := getOffset(idx, idx.Offset32[i][k*4:(k+1)*4])
			objs[l] = append(objs[l], ObjEntry{offset, offmap[offset]})
		}
	}
*/
// 	offmap := make(map[uint64]uint64)
