package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"strconv"
	"time"

	"github.com/eliquious/primes/dproc"
	"github.com/kavehmz/prime"
)

// TypePrime is the message type for prime numbers
const TypePrime = dproc.MessageType("Prime")

// PrimeMessage stores the prime and its index.
type PrimeMessage struct {
	Index int
	Prime uint64
}

// PrimeGenerator generates primes below a given limit.
type PrimeGenerator struct {
	Limit uint64
}

// Handle sends prime messages to all the child processes.
func (p *PrimeGenerator) Handle(ctx context.Context, proc dproc.Process, msg dproc.Message) {
	switch msg.Type {
	default:
		fmt.Println("Unknown message type: ", msg.Type)
	case dproc.MessageTypeStart:
		log.Printf("[%s] - Starting...", proc.Name())
		log.Printf("[%s] - Finding primes...", proc.Name())
		primes := prime.Primes(p.Limit)
		log.Printf("[%s] - Found %d primes", proc.Name(), len(primes))

		numPrimes := len(primes)
		for i, p := range primes {
			if i%(numPrimes/10) == 0 {
				log.Printf("[%s] - Processing primes - %.0f%%", proc.Name(), float64(i)/float64(numPrimes)*100)
			}

			proc.Children().Dispatch(dproc.Message{
				Forward:   false,
				Type:      TypePrime,
				Timestamp: time.Now().UTC(),
				Value: PrimeMessage{
					Index: i,
					Prime: p,
				},
			})
		}
		proc.SetState(dproc.StateKilled)
		log.Printf("[%s] - Exiting...", proc.Name())
	}
}

// TypeExtendedPrime is the message type for extended prime numbers
const TypeExtendedPrime = dproc.MessageType("ExtendedPrime")

// ExtendedPrimeMessage extends the prime message with additional fields.
type ExtendedPrimeMessage struct {
	Index        int
	Prime        uint64
	PrimeDiff    uint64
	LastDigit    rune
	Log          float64
	X            float64
	Y            float64
	Azimuth      float64
	DeltaAzimuth float64
	Mod360       uint64
	Revs         uint64
}

// PrimeExtensionProcess adds data to a prime before passing it on
type PrimeExtensionProcess struct {
	PreviousPrime PrimeMessage
	PreviousAzm   float64
}

// Handle sends prime messages to all the child processes.
func (p *PrimeExtensionProcess) Handle(ctx context.Context, proc dproc.Process, msg dproc.Message) {
	switch msg.Type {
	default:
		log.Printf("[%s] - Unknown message type: %s", proc.Name(), msg.Type)
	case dproc.MessageTypeStart:
		log.Printf("[%s] - Starting...", proc.Name())
	case dproc.MessageTypeStop:
		log.Printf("[%s] - Exiting...", proc.Name())
	case TypePrime:
		msg := msg.Value.(PrimeMessage)
		// pRad := float64(msg.Prime)
		pRad := float64(msg.Prime) / 180. * math.Pi
		pLog := math.Log10(float64(msg.Prime))
		pX := pLog * math.Cos(pRad)
		pY := pLog * math.Sin(pRad)

		pAzm := math.Atan2(pY, pX) * 180 / math.Pi
		if pAzm < 0 {
			pAzm += 360
		}

		deltaAngle := pAzm - p.PreviousAzm
		if deltaAngle < 0 {
			deltaAngle += 360
		}
		proc.Children().Dispatch(dproc.Message{
			Forward:   false,
			Type:      TypeExtendedPrime,
			Timestamp: time.Now().UTC(),
			Value: ExtendedPrimeMessage{
				Index:        msg.Index,
				Prime:        msg.Prime,
				PrimeDiff:    msg.Prime - p.PreviousPrime.Prime,
				LastDigit:    lastDigit(msg.Prime),
				Log:          pLog,
				X:            pX,
				Y:            pY,
				Azimuth:      pAzm,
				DeltaAzimuth: deltaAngle,
				Mod360:       msg.Prime % 360,
				Revs:         msg.Prime / 360,
			},
		})
		p.PreviousPrime = msg
		p.PreviousAzm = pAzm
	}
}

func lastDigit(p uint64) rune {
	s := []rune(strconv.FormatUint(p, 10))
	return s[len(s)-1] - rune(48)
}

func sumOfString(s string) (sum uint64, last uint64) {
	for _, c := range s {
		last = uint64(c) - 48
		sum += last
	}
	return sum, last
}

// TypeMapKey is the message type for map results
const TypeMapKey = dproc.MessageType("MapKey")

// ExtendedPrimeMapperFunc maps ExtendedPrimeMessages to strings
type ExtendedPrimeMapperFunc func(msg ExtendedPrimeMessage) string

// PrimeAzimuthMapperFunc maps ExtendedPrimeMessages to their azimuth
func PrimeAzimuthMapperFunc(msg ExtendedPrimeMessage) string {
	return fmt.Sprintf("%.f", msg.Azimuth)
}

// LastDigitPrimeAzimuthMapperFunc maps ExtendedPrimeMessages to their azimuth and last digit
func LastDigitPrimeAzimuthMapperFunc(msg ExtendedPrimeMessage) string {
	return fmt.Sprintf("%d,%.f", msg.LastDigit, msg.Azimuth)
}

// LastDigitPrimeAzimuthDeltaMapperFunc maps ExtendedPrimeMessages to their last digit, azimuth and gap.
func LastDigitPrimeAzimuthDeltaMapperFunc(msg ExtendedPrimeMessage) string {
	return fmt.Sprintf("%d,%.f,%d", msg.LastDigit, msg.Azimuth, msg.PrimeDiff)
}

// LastDigitPrimeDeltaMapperFunc maps ExtendedPrimeMessages to their last digit and gap.
func LastDigitPrimeDeltaMapperFunc(msg ExtendedPrimeMessage) string {
	return fmt.Sprintf("%d,%d", msg.LastDigit, msg.PrimeDiff)
}

// DeltaLastDigitAzimuthMapperFunc maps ExtendedPrimeMessages to their gap, last digit and azimuth.
func DeltaLastDigitAzimuthMapperFunc(msg ExtendedPrimeMessage) string {
	return fmt.Sprintf("% 3d,%d,%.f", msg.PrimeDiff, msg.LastDigit, msg.Azimuth)
}

// RevMapperFunc maps ExtendedPrimeMessages to their revolution.
func RevMapperFunc(msg ExtendedPrimeMessage) string {
	return fmt.Sprintf("%04d", msg.Revs)
}

// LastDigitAzimuthPreviousAzimuthMapperFunc maps ExtendedPrimeMessages to their gap, last digit and azimuth.
func LastDigitAzimuthPreviousAzimuthMapperFunc(msg ExtendedPrimeMessage) string {
	prevAzm := msg.Azimuth - msg.DeltaAzimuth
	// if prevAzm < 0 {
	// 	prevAzm += 360
	// }
	return fmt.Sprintf("% 3d,%.f,%.f", msg.LastDigit, msg.Azimuth, prevAzm)
}

// PrimeMapperProcess maps data
type PrimeMapperProcess struct {
	MapperFunc ExtendedPrimeMapperFunc
}

// Handle sends prime messages to all the child processes.
func (p *PrimeMapperProcess) Handle(ctx context.Context, proc dproc.Process, msg dproc.Message) {
	switch msg.Type {
	default:
		log.Printf("[%s] - Unknown message type: %s", proc.Name(), msg.Type)
	case dproc.MessageTypeStart:
		log.Printf("[%s] - Starting...", proc.Name())
	case dproc.MessageTypeStop:
		log.Printf("[%s] - Exiting...", proc.Name())
	case TypeExtendedPrime:
		msg := msg.Value.(ExtendedPrimeMessage)
		key := p.MapperFunc(msg)

		proc.Children().Dispatch(dproc.Message{
			Forward:   false,
			Type:      TypeMapKey,
			Timestamp: time.Now().UTC(),
			Value:     key,
		})
	}
}
