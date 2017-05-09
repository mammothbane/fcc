package main

import (
	"fmt"
	"log"

	"time"

	"github.com/mammothbane/fcc"
)

type backoff struct {
	factor  float64
	current time.Duration
	max     time.Duration
}

func (b *backoff) do(op func() *fcc.Error) *fcc.Error {
	for {
		err := op()
		if err == nil || err.Fatal() {
			return err
		}

		b.current = time.Duration(float64(b.current) * b.factor)
		if b.current > b.max {
			return err
		}

		time.Sleep(b.current)
	}
}

func main() {
	bck := &backoff{
		factor:  2,
		current: 10 * time.Millisecond,
		max:     10 * time.Second,
	}
	var proc *fcc.Proc

	log.Println("retrieving proceeding...")
	err := bck.do(func() *fcc.Error {
		inner, err := fcc.Proceeding("17-108")
		if err == nil {
			proc = inner
		} else if !err.Fatal() {
			log.Printf("Error: '%v'. Retrying.", err)
		}

		return err
	})

	if err != nil {
		log.Fatal(err)
	}

	// fill with your information.
	info := fcc.FilingInfo{}

	var conf *fcc.FilingConfirmation

	log.Println("submitting filing info...")
	bck.do(func() *fcc.Error {
		inner, err := info.Submit(proc)
		if err == nil {
			conf = inner
		} else if !err.Fatal() {
			log.Printf("Error: '%v'. Retrying.", err)
		}

		return err
	})

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Submission successful:", conf)
}
