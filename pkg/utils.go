package game

import (
	"bytes"
	"encoding/gob"
	"log"
)

func ToGob[T any](from T) []byte {
	buf := new(bytes.Buffer)
	encoder := gob.NewEncoder(buf)
	err := encoder.Encode(from)
	if err != nil {
		log.Printf("can't convert to gob: %s", err.Error())
	}
	return buf.Bytes()

}

func FromGob[T any](from []byte, to *T) {
	buf := bytes.NewBuffer(from)
	decoder := gob.NewDecoder(buf)
	err := decoder.Decode(to)
	if err != nil {
		log.Printf("can't convert from gob: %s", err.Error())
	}
}

// ReverseStrings reverses a slice of strings in place
func ReverseStrings(xs []string) {
	for i := 0; i < len(xs)/2; i++ {
		xs[i], xs[len(xs)-1-i] = xs[len(xs)-1-i], xs[i]
	}
}
