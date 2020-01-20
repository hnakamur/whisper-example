package main

import (
	"bufio"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mattn/go-zglob"
)

type metadata struct {
	aggType        uint32
	maxRetension   uint32
	xFilesFactor   uint32
	retentionCount uint32
}

func (m *metadata) readFrom(r io.Reader) error {
	err := binary.Read(r, binary.BigEndian, &m.aggType)
	if err != nil {
		return err
	}
	err = binary.Read(r, binary.BigEndian, &m.maxRetension)
	if err != nil {
		return err
	}
	err = binary.Read(r, binary.BigEndian, &m.xFilesFactor)
	if err != nil {
		return err
	}
	return binary.Read(r, binary.BigEndian, &m.retentionCount)
}

type retention struct {
	offset          uint32
	secondsPerPoint uint32
	numberOfPoints  uint32
}

func (rt *retention) readFrom(r io.Reader) error {
	err := binary.Read(r, binary.BigEndian, &rt.offset)
	if err != nil {
		return err
	}
	err = binary.Read(r, binary.BigEndian, &rt.secondsPerPoint)
	if err != nil {
		return err
	}
	return binary.Read(r, binary.BigEndian, &rt.numberOfPoints)
}

type dataPoint struct {
	interval uint32
	value    float64
}

func (p *dataPoint) readFrom(r io.Reader) error {
	err := binary.Read(r, binary.BigEndian, &p.interval)
	if err != nil {
		return err
	}
	v, err := readFloat64From(r)
	if err != nil {
		return err
	}
	p.value = v
	return nil
}

func readFloat64From(r io.Reader) (float64, error) {
	var intVal uint64
	err := binary.Read(r, binary.BigEndian, &intVal)
	if err != nil {
		return math.NaN(), err
	}
	return math.Float64frombits(intVal), nil
}

func getMetricNameFromRelPath(relPath string) string {
	if strings.HasSuffix(relPath, ".wsp") {
		relPath = relPath[:len(relPath)-len(".wsp")]
	}
	return strings.ReplaceAll(relPath, "/", ".")
}

func printLatestValue(whisperDir, filename string) error {
	relPath, err := filepath.Rel(whisperDir, filename)
	if err != nil {
		return err
	}
	metricName := getMetricNameFromRelPath(relPath)

	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	r := bufio.NewReader(file)

	m := &metadata{}
	err = m.readFrom(r)
	if err != nil {
		return err
	}

	retentions := make([]retention, m.retentionCount)
	for i := 0; i < len(retentions); i++ {
		err = retentions[i].readFrom(r)
		if err != nil {
			return err
		}
	}

	var latest, p dataPoint
	n := retentions[0].numberOfPoints
	for i := 0; i < int(n); i++ {
		if err := p.readFrom(r); err != nil {
			return err
		}
		if p.interval > latest.interval {
			latest = p
		}
	}

	fmt.Printf("%d,%s,%.0f\n", latest.interval, metricName, latest.value)
	return nil
}

func run(whisperDir, globPattern string, interval, offset time.Duration) error {
	for {
		// See https://github.com/grafana/carbon-relay-ng/blob/76e4304680ab9c6b186084746b8b2589fd9d8849/clock/clock.go#L23-L26
		unix := time.Now().UnixNano()
		adjusted := time.Duration(unix) - offset
		diff := time.Duration(interval - adjusted%interval)
		log.Printf("sleep %s", diff)
		time.Sleep(diff)

		filenames, err := zglob.Glob(filepath.Join(whisperDir, globPattern))
		if err != nil {
			return err
		}
		log.Printf("reading %d whisper files", len(filenames))
		for _, filename := range filenames {
			if err := printLatestValue(whisperDir, filename); err != nil {
				return err
			}
		}
	}
}

func main() {
	whisperDir := flag.String("whisper-dir", "", "whisper file directory (required)")
	globPattern := flag.String("glob", "", "whisper file glob pattern (required)")
	interval := flag.Duration("interval", time.Minute, "watch interval")
	offset := flag.Duration("offset", 5*time.Second, "watch offset")
	flag.Parse()
	if *whisperDir == "" || *globPattern == "" {
		flag.Usage()
		os.Exit(1)
	}

	err := run(*whisperDir, *globPattern, *interval, *offset)
	if err != nil {
		log.Fatal(err)
	}
}
