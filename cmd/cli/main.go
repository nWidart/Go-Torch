package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"GoTorch/internal/parser"
	"GoTorch/internal/tailer"
	"GoTorch/internal/tracker"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

// run is the testable entrypoint for the CLI. It returns an exit code rather than exiting directly.
func run(args []string) int {
	fs := flag.NewFlagSet("cli", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	logPath := fs.String("log", "", "Path to Torchlight Infinite log file")
	fromStart := fs.Bool("from-start", false, "Read from start instead of tailing from end")
	pollMs := fs.Int("poll-ms", 300, "Polling interval in milliseconds")
	debug := fs.Bool("debug", true, "Print parsed events and errors")
	once := fs.Bool("once", false, "Process the file once and exit (no live tail)")
	if err := fs.Parse(args); err != nil {
		fmt.Println("Usage: cli --log <path> [--from-start] [--poll-ms N] [--debug] [--once]")
		return 2
	}

	if *logPath == "" {
		fmt.Println("Usage: cli --log <path> [--from-start] [--poll-ms N] [--debug] [--once]")
		return 2
	}

	// Resolve env vars in path (Windows %USERPROFILE%, etc.)
	resolved := os.ExpandEnv(*logPath)
	if resolved != *logPath {
		*logPath = resolved
	}

	p := parser.New()
	trk := tracker.New()

	if *once {
		if err := processOnce(*logPath, p, trk, *debug); err != nil {
			fmt.Println("error:", err)
			return 1
		}
		printState(trk)
		return 0
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle Ctrl+C
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nStopping...")
		cancel()
	}()

	lines := make(chan string, 1024)
	t := tailer.New(tailer.Options{Path: *logPath, FromStart: *fromStart, PollEvery: time.Duration(*pollMs) * time.Millisecond})

	go func() {
		if err := t.Start(ctx, lines); err != nil {
			if *debug {
				fmt.Println("tailer error:", err)
			}
		}
	}()

	lastPrint := time.Now()
	scanner := bufio.NewScanner(readerFromChan(ctx, lines))
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if ev := p.Parse(line); ev != nil {
			trk.OnEvent(ev)
			if *debug {
				fmt.Printf("[%s] %s\n", ev.Time.Format(time.Kitchen), ev.Kind)
			}
		}
		if time.Since(lastPrint) >= 1*time.Second {
			printState(trk)
			lastPrint = time.Now()
		}
	}
	if err := scanner.Err(); err != nil && ctx.Err() == nil {
		fmt.Println("scanner error:", err)
	}
	return 0
}

func processOnce(path string, p *parser.Parser, trk *tracker.Tracker, debug bool) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	s.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for s.Scan() {
		line := s.Text()
		if ev := p.Parse(line); ev != nil {
			trk.OnEvent(ev)
			if debug {
				fmt.Printf("[%s] %s\n", ev.Time.Format(time.Kitchen), ev.Kind)
			}
		}
	}
	return s.Err()
}

func readerFromChan(ctx context.Context, ch <-chan string) *chanReader {
	return &chanReader{ctx: ctx, ch: ch}
}

type chanReader struct {
	ctx context.Context
	ch  <-chan string
}

func (r *chanReader) Read(p []byte) (int, error) {
	select {
	case <-r.ctx.Done():
		return 0, r.ctx.Err()
	case s, ok := <-r.ch:
		if !ok {
			return 0, context.Canceled
		}
		copy(p, s)
		if len(p) > len(s) {
			p[len(s)] = '\n'
			return len(s) + 1, nil
		}
		return len(s), nil
	}
}

func printState(trk *tracker.Tracker) {
	st := trk.GetState()
	status := "Idle"
	if st.InMap && st.Current.Active {
		status = "In Map"
	}
	dur := time.Since(st.Current.StartedAt)
	if !st.Current.Active {
		dur = st.Current.EndedAt.Sub(st.Current.StartedAt)
	}
	if dur < 0 {
		dur = 0
	}
	fmt.Println("------------------------------")
	fmt.Printf("Status: %s\n", status)
	if st.Current.StartedAt.IsZero() {
		fmt.Println("No session yet.")
		return
	}
	fmt.Printf("Session: %s\n", st.Current.StartedAt.Format(time.RFC3339))
	fmt.Printf("Duration: %s\n", dur.Truncate(time.Second))

	// Totals this session by ConfigBaseID
	if len(st.Current.Tally) == 0 {
		fmt.Println("Tally: (none yet)")
	} else {
		pairs := make([]string, 0, len(st.Current.Tally))
		var total int
		for id, n := range st.Current.Tally {
			pairs = append(pairs, fmt.Sprintf("%d=%d", id, n))
			total += n
		}
		fmt.Println("Tally:", strings.Join(pairs, ", "))
		fmt.Printf("Session total: %d\n", total)
		if dur > 0 {
			h := dur.Hours()
			iph := float64(total) / h
			fmt.Printf("Items/hour: %.1f\n", iph)
		}
	}

	// Last few events
	if len(st.LastEvents) > 0 {
		last := st.LastEvents
		if len(last) > 10 {
			last = last[len(last)-10:]
		}
		fmt.Println("Recent events:")
		for _, ev := range last {
			fmt.Printf("- %s %s\n", ev.Time.Format(time.Kitchen), ev.Kind)
		}
	}
}
