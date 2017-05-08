package rif_17108

import (
	"log"
)

func main() {
	proc, err := proceeding("rif_17108")
	if err != nil {
		log.Fatal(err)
	}
}
