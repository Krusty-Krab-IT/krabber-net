package ksuid

import (
	cr "crypto/rand"
	"encoding/binary"
	"fmt"
	"github.com/segmentio/ksuid"
	"time"
)

func GenerateKSUID() ksuid.KSUID {
	b := make([]byte, 16)
	r, err := cr.Read(b) // random
	if err != nil {
		fmt.Printf("ERR %s", err)
	}
	p := toByteArray(r)
	bar := p[:]
	t := time.Unix(time.Now().Unix(), 0)
	k, err := ksuid.FromParts(t, bar)
	if err != nil {
		fmt.Printf("err: %s", err)
	}
	return k
}

func toByteArray(i int) (arr [16]byte) {
	binary.BigEndian.PutUint32(arr[0:16], uint32(i))
	return
}
