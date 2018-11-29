package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/eliquious/primes/dproc"
)

func main() {
	log.SetOutput(os.Stdout)
	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	engine := dproc.NewEngine(ctx, cancel, dproc.ProcessList{
		dproc.NewDefaultProcess(ctx, "Prime Process", &PrimeGenerator{1e5}, dproc.ProcessList{
			dproc.NewDefaultProcess(ctx, "Prime Extension", &PrimeExtensionProcess{}, dproc.ProcessList{
				dproc.NewDefaultProcess(ctx, "Prime Output", NewPrimeOutput("primes.csv"), dproc.ProcessList{}),
				dproc.NewDefaultProcess(ctx, "Rev Size Prime Output", NewPrimeRevSizeOutput("rev_size.csv"), dproc.ProcessList{}),
				dproc.NewDefaultProcess(ctx, "Azimuth Mapper", &PrimeMapperProcess{PrimeAzimuthMapperFunc}, dproc.ProcessList{
					dproc.NewDefaultProcess(ctx, "Azimuth Mapper Output", NewPrimeMapperOutput("azm_map_results.csv", "azm"), dproc.ProcessList{}),
				}),
				dproc.NewDefaultProcess(ctx, "Last Digit, Azimuth Mapper", &PrimeMapperProcess{LastDigitPrimeAzimuthMapperFunc}, dproc.ProcessList{
					dproc.NewDefaultProcess(ctx, "Last Digit Azimuth Mapper Output", NewPrimeMapperOutput("last_digit_azm_map_results.csv", "lastDigit,azm"), dproc.ProcessList{}),
				}),
				dproc.NewDefaultProcess(ctx, "Last Digit, Azimuth, Delta Mapper", &PrimeMapperProcess{LastDigitPrimeAzimuthDeltaMapperFunc}, dproc.ProcessList{
					dproc.NewDefaultProcess(ctx, "Last Digit Azimuth Delta Mapper Output", NewPrimeMapperOutput("last_digit_azm_delta_map_results.csv", "lastDigit,azm,delta"), dproc.ProcessList{}),
				}),
				dproc.NewDefaultProcess(ctx, "Last Digit, Delta Mapper", &PrimeMapperProcess{LastDigitPrimeDeltaMapperFunc}, dproc.ProcessList{
					dproc.NewDefaultProcess(ctx, "Last Digit Delta Mapper Output", NewPrimeMapperOutput("last_digit_delta_map_results.csv", "lastDigit,delta"), dproc.ProcessList{}),
				}),
				dproc.NewDefaultProcess(ctx, "Delta, Last Digit, Azimuth Mapper", &PrimeMapperProcess{DeltaLastDigitAzimuthMapperFunc}, dproc.ProcessList{
					dproc.NewDefaultProcess(ctx, "Delta Last Digit Azimuth Mapper Output", NewPrimeMapperOutput("delta_last_digit_azm_map_results.csv", "delta,lastDigit,azm"), dproc.ProcessList{}),
				}),
				dproc.NewDefaultProcess(ctx, "Rev Mapper", &PrimeMapperProcess{RevMapperFunc}, dproc.ProcessList{
					dproc.NewDefaultProcess(ctx, "Rev Mapper Output", NewPrimeMapperOutput("rev_map_results.csv", "rev"), dproc.ProcessList{}),
				}),
				dproc.NewDefaultProcess(ctx, "Last Digit, Azimuth, Prev Azimuth Mapper", &PrimeMapperProcess{LastDigitAzimuthPreviousAzimuthMapperFunc}, dproc.ProcessList{
					dproc.NewDefaultProcess(ctx, "Last Digit Azimuth Prev Azimuth Mapper Output", NewPrimeMapperOutput("last_digit_azm_prevazm_map_results.csv", "lastDigit,azm,prevAzm"), dproc.ProcessList{}),
				}),
			}),
		}),
	})
	start := time.Now()
	engine.Start(&wg)

	wg.Wait()
	fmt.Println("Elapsed: ", time.Since(start))
	fmt.Println("\nExiting...")
}
