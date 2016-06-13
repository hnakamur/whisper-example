package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"time"

	"github.com/lomik/go-whisper"
)

func readWhisperFile(filename string) error {
	w, err := whisper.Open(filename)
	if err != nil {
		return err
	}
	defer w.Close()

	now := time.Now()
	maxRetension := time.Duration(w.MaxRetention()) * time.Second
	untilTime := now
	fromTime := untilTime.Add(-maxRetension)
	log.Printf("fromTime=%v", fromTime)
	log.Printf("untilTime=%v", untilTime)
	series, err := w.Fetch(int(fromTime.Unix()), int(untilTime.Unix()))
	if err != nil {
		return err
	}
	if series != nil {
		log.Printf("series.fromTime=%v", time.Unix(int64(series.FromTime()), 0))
		log.Printf("series.untilTime=%v", time.Unix(int64(series.UntilTime()), 0))
		for _, p := range series.Points() {
			if !math.IsNaN(p.Value) {
				log.Printf("time=%d (%v), value=%g", p.Time, time.Unix(int64(p.Time), 0).In(time.UTC).Format(time.RFC3339Nano), p.Value)
			}
		}
		log.Printf("step=%d", series.Step())
	}

	return nil
}

func main() {
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Println("Usage: whisper-example wspfilename")
		os.Exit(1)
	}
	filename := flag.Arg(0)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	var err error

	err = readWhisperFile(filename)
	if err != nil {
		log.Fatal(err)
	}
}
