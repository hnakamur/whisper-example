package main

import (
	"flag"
	"fmt"
	"log"
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
	untilTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	fromTime := untilTime.Add(-7 * 24 * time.Hour)
	log.Printf("fromTime=%v", fromTime)
	log.Printf("untilTime=%v", untilTime)
	series, err := w.Fetch(int(fromTime.Unix()), int(untilTime.Unix()))
	if err != nil {
		return err
	}
	log.Printf("series=%v", *series)
	log.Printf("series.fromTime=%v", time.Unix(int64(series.FromTime()), 0))
	log.Printf("series.untilTime=%v", time.Unix(int64(series.UntilTime()), 0))
	log.Printf("step=%d", series.Step())

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
