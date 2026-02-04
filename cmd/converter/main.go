package main

import (
	"io"
	"os"

	"github.com/freehandle/breeze/util"
)

func main() {
	file, err := os.Open("../proxy-handles/blocos_old")
	if err != nil {
		panic(err)

	}
	defer file.Close()
	saida, err := os.Create("../proxy-handles/blocos_0")
	if err != nil {
		panic(err)
	}
	defer saida.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}
	pos := 0
	// epoca := uint64(0)
	nactions := uint32(0)
	tamanho := uint32(0)
	bytes := make([]byte, 0)
	for {
		_, pos = util.ParseUint64(data, pos)
		nactions, pos = util.ParseUint32(data, pos)
		// pegou o epoca e o nactions
		bytes = append(bytes, data[pos-12:pos]...)
		if nactions > 0 {
			for range int(nactions) {
				tamanho, pos = util.ParseUint32(data, pos)
				util.PutUint32(tamanho-1, &bytes)
				bytes = append(bytes, data[pos+1:pos+int(tamanho)]...)
				pos += int(tamanho)
			}
		}
		if pos == len(data) {
			break
		}
	}
	n, err := saida.Write(bytes)
	if err != nil {
		panic(err)
	}
	if n != len(bytes) {
		panic("erro de tamanho")
	}
}
