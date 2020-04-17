package main

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/a4a881d4/gitcrawling/packext"
	"gopkg.in/src-d/go-git.v4/plumbing/format/packfile"

	"github.com/a4a881d4/gitcrawling/types"
)

func main() {
	var s = make([]byte, 256)
	for k, _ := range s {
		s[k] = byte(k)
	}
	pn, err := packext.NewPackNo(s, os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	defer pn.Close()
	test, terr := hex.DecodeString(os.Args[2])
	var key types.Hash
	if terr == nil {
		copy(key[:], test[:])
	}
	body := GetSome(pn, key)
	fmt.Println(string(body))
}

func GetSome(pn *packext.Packns, k types.Hash) []byte {
	raw, deta, base, err := pn.GetAnyHash(k)
	if err != nil {
		fmt.Println(err)
		return []byte{}
	}
	if deta {
		next, err := GetSpecial(pn, base)
		if err != nil {
			fmt.Println(err)
			return []byte{}
		}
		patched, err := packfile.PatchDelta(next, raw)
		if err != nil {
			fmt.Println(err)
			return []byte{}
		} else {
			return patched
		}
	} else {
		return raw
	}
}

func GetSpecial(pn *packext.Packns, h [32]byte) (body []byte, err error) {
	var raw, next []byte
	var deta bool
	var base [32]byte

	raw, deta, base, err = pn.GetSpecial(h)
	if err != nil {
		return
	}
	if deta {
		next, err = GetSpecial(pn, base)
		if err != nil {
			return
		}
		body, err = packfile.PatchDelta(next, raw)
		if err != nil {
			fmt.Println(string(raw))
			fmt.Println(string(next))
			return
		}

	} else {
		body = raw
		return
	}
	return
}
