package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/eliquious/primes/dproc"
)

// NewPrimeOutput creates a new PrimeOutput.
func NewPrimeOutput(filename string) *PrimeOutput {
	output, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Fatal(err)
	}
	output.Truncate(0)
	return &PrimeOutput{output}
}

// PrimeOutput writes a log entry for each prime
type PrimeOutput struct {
	Writer *os.File
}

// Handle writes primes info to file.
func (o *PrimeOutput) Handle(ctx context.Context, proc dproc.Process, msg dproc.Message) {
	switch msg.Type {
	default:
		log.Printf("[%s] - Unknown message type: %s", proc.Name(), msg.Type)
	case dproc.MessageTypeStart:
		log.Printf("[%s] - Starting...", proc.Name())
		log.Printf("[%s] - Writing output header...", proc.Name())
		o.Writer.WriteString("index,prime,diff,lastDigit,pLog,pX,pY,pAzm,deltaAzm,mod360,revs\n")
	case TypeExtendedPrime:
		extMsg := msg.Value.(ExtendedPrimeMessage)
		o.Writer.WriteString(fmt.Sprintf("%d,%d,%d,%d,%.5f,%.5f,%.5f,%.f,%.f,%d,%d\n",
			extMsg.Index,
			extMsg.Prime,
			extMsg.PrimeDiff,
			extMsg.LastDigit,
			extMsg.Log,
			extMsg.X,
			extMsg.Y,
			extMsg.Azimuth,
			extMsg.DeltaAzimuth,
			extMsg.Mod360,
			extMsg.Revs,
		))
	case dproc.MessageTypeStop:
		o.Writer.Close()
		log.Printf("[%s] - Exiting...", proc.Name())
	}
}

// NewPrimeMapperOutput creates a new PrimeMapperOutput.
func NewPrimeMapperOutput(filename, keyHeader string) *PrimeMapperOutput {
	output, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Fatal(err)
	}
	output.Truncate(0)
	return &PrimeMapperOutput{keyHeader, output, make(map[string]int)}
}

// PrimeMapperOutput writes an entry for each mapper key and count
type PrimeMapperOutput struct {
	KeyHeader string
	Writer    *os.File
	Data      map[string]int
}

// Handle counts map results then writes data to file.
func (o *PrimeMapperOutput) Handle(ctx context.Context, proc dproc.Process, msg dproc.Message) {
	switch msg.Type {
	default:
		log.Printf("[%s] - Unknown message type: %s", proc.Name(), msg.Type)
	case dproc.MessageTypeStart:
		log.Printf("[%s] - Starting...", proc.Name())
		log.Printf("[%s] - Writing output header...", proc.Name())
		o.Writer.WriteString(o.KeyHeader + ",count\n")
	case TypeMapKey:
		key := msg.Value.(string)
		o.Data[key]++
	case dproc.MessageTypeStop:
		lines := make([]string, len(o.Data))
		for k, v := range o.Data {
			lines = append(lines, fmt.Sprintf("%s,%d\n", k, v))
		}
		sort.Strings(lines)
		o.Writer.WriteString(strings.Join(lines, ""))

		o.Writer.Close()
		log.Printf("[%s] - Exiting...", proc.Name())
	}
}

// NewPrimeRevSizeOutput creates a new PrimeRevSizeOutput.
func NewPrimeRevSizeOutput(filename string) *PrimeRevSizeOutput {
	output, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Fatal(err)
	}
	output.Truncate(0)
	return &PrimeRevSizeOutput{output, 2, 0, 2, 0, 2.}
}

// PrimeRevSizeOutput writes a log entry for each prime revolution
type PrimeRevSizeOutput struct {
	Writer        *os.File
	FirstRevPrime uint64
	LastRev       uint64
	LastPrime     uint64
	RevCount      int
	RevStartAzm   float64
}

// Handle writes primes info to file.
func (o *PrimeRevSizeOutput) Handle(ctx context.Context, proc dproc.Process, msg dproc.Message) {
	switch msg.Type {
	default:
		log.Printf("[%s] - Unknown message type: %s", proc.Name(), msg.Type)
	case dproc.MessageTypeStart:
		log.Printf("[%s] - Starting...", proc.Name())
		log.Printf("[%s] - Writing output header...", proc.Name())
		o.Writer.WriteString("rev,start,span,azm,count\n")
	case TypeExtendedPrime:
		extMsg := msg.Value.(ExtendedPrimeMessage)

		// New revolution
		if extMsg.Revs != o.LastRev {

			o.Writer.WriteString(fmt.Sprintf("%d,%d,%d,%.f,%d\n",
				o.LastRev,
				o.FirstRevPrime,
				o.LastPrime-o.FirstRevPrime,
				o.RevStartAzm,
				o.RevCount,
			))
			o.FirstRevPrime = extMsg.Prime
			o.RevStartAzm = extMsg.Azimuth
			o.LastRev = extMsg.Revs
			o.RevCount = 0
		}
		o.LastPrime = extMsg.Prime
		o.RevCount++
	case dproc.MessageTypeStop:
		o.Writer.WriteString(fmt.Sprintf("%d,%d,%d,%.f,%d\n",
			o.LastRev,
			o.FirstRevPrime,
			o.LastPrime-o.FirstRevPrime,
			o.RevStartAzm,
			o.RevCount,
		))
		o.Writer.Close()
		log.Printf("[%s] - Exiting...", proc.Name())
	}
}
