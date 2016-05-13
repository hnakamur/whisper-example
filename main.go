package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/lomik/go-whisper"
	"github.com/marpaia/graphite-golang"
)

func sendToGraphite(graphiteAddr string, graphitePort int, t time.Time, item string, value float64) error {
	g, err := graphite.NewGraphite(graphiteAddr, graphitePort)
	if err != nil {
		return err
	}
	defer g.Disconnect()

	metric := graphite.Metric{
		Name:      item,
		Value:     strconv.FormatFloat(value, 'g', -1, 64),
		Timestamp: t.Unix(),
	}
	return g.SendMetric(metric)
}

func readWhisperFile(filename string) error {
	w, err := whisper.Open(filename)
	if err != nil {
		return err
	}
	defer w.Close()

	log.Printf("w=%v", w)
	startTimeSec := w.StartTime()
	startTime := time.Unix(int64(startTimeSec), 0)
	log.Printf("startTime=%v (%d)", startTime, startTimeSec)
	log.Printf("aggregationMethod=%s", w.AggregationMethod())
	log.Printf("maxRetention=%d", w.MaxRetention())
	log.Printf("size=%d", w.Size())
	log.Printf("xFilesFactor=%f", w.XFilesFactor())
	retentions := w.Retentions()
	for i, retention := range retentions {
		log.Printf("i=%d, retention=%v", i, retention)
	}

	now := time.Now()
	untilTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	//untilTime = untilTime.Add(-time.Hour)
	//.Add(-5 * time.Minute)
	//fromTime := untilTime.Add(-4 * time.Hour)
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
	// 2016/05/10 15:57:37.245843 now=2016-05-10 15:57:37.245734652 +0900 JST
	/*
		now := time.Now()
			log.Printf("now=%v", now)
			err = sendToGraphite("127.0.0.1", 2003, now, "foo", 3)
			if err != nil {
				log.Fatal(err)
			}
	*/

	err = readWhisperFile(filename)
	if err != nil {
		log.Fatal(err)
	}
}
