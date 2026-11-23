// pir-preprocess-mpi is the explicit experimental distributed GPU server entry.
package main

import (
	"crypto/rand"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	mathrand "math/rand"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/pprof"
	"sync"
	"syscall"
	"time"

	lrs "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-shared"
)

func readConfig(path string) (lrs.DistributedPreprocessingConfig, error) {
	f, e := os.Open(path)
	if e != nil {
		return lrs.DistributedPreprocessingConfig{}, e
	}
	defer f.Close()
	return lrs.ReadDistributedPreprocessingConfig(f)
}

// Guard observations are not progress. Arithmetic emits completed tile/pass work.
func resources(c lrs.DistributedPreprocessingConfig, began time.Time) error {
	if time.Since(began) > time.Duration(c.DeadlineSeconds)*time.Second {
		return fmt.Errorf("rank deadline")
	}
	var usage syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &usage); err != nil {
		return err
	}
	if uint64(usage.Maxrss)*1024 > c.RSSBytes {
		return fmt.Errorf("rank RSS budget")
	}
	var total uint64
	err := filepath.WalkDir(c.Directory, func(path string, d os.DirEntry, err error) error {
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			return err
		}
		if d.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlink in owned output")
		}
		if d.IsDir() {
			return nil
		}
		info, e := d.Info()
		if os.IsNotExist(e) {
			return nil
		}
		if e != nil {
			return e
		}
		if info.Size() < 0 || uint64(info.Size()) > c.FileBytes {
			return fmt.Errorf("file budget")
		}
		total += uint64(info.Size())
		if total > c.StorageBytes {
			return fmt.Errorf("output storage budget")
		}
		return nil
	})
	return err
}

func run(args []string) error {
	flags := flag.NewFlagSet("pir-preprocess-mpi", flag.ContinueOnError)
	configPath := flags.String("config", "", "required explicit JSON job policy")
	profilePath := flags.String("profile", "", "required fresh per-rank CPU profile")
	check := flags.Bool("check-config", false, "validate scalar policy without MPI/CUDA")
	testSeed := flags.Int64("test-seed", 0, "explicit reproducible validation randomness;omit for secure source")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected arguments")
	}
	if *configPath == "" {
		return fmt.Errorf("--config is required")
	}
	c, err := readConfig(*configPath)
	if err != nil {
		return err
	}
	if *check {
		return json.NewEncoder(os.Stdout).Encode(c)
	}
	if *profilePath == "" {
		return fmt.Errorf("fresh per-rank profile required")
	}
	info, e := os.Stat(c.Library)
	if e != nil || !info.Mode().IsRegular() {
		return fmt.Errorf("CUDA library path preflight: %v", e)
	}
	probe, e := os.CreateTemp(c.Directory, ".rank-preflight-")
	if e != nil {
		return e
	}
	name := probe.Name()
	ce := probe.Close()
	re := os.Remove(name)
	if ce != nil {
		return ce
	}
	if re != nil {
		return re
	}
	if err = resources(c, time.Now()); err != nil {
		return err
	}
	if err = syscall.Setrlimit(syscall.RLIMIT_FSIZE, &syscall.Rlimit{Cur: c.FileBytes, Max: c.FileBytes}); err != nil {
		return err
	}
	profile, err := os.OpenFile(*profilePath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	if err = pprof.StartCPUProfile(profile); err != nil {
		_ = profile.Close()
		return err
	}
	var once sync.Once
	stopProfile := func() { once.Do(func() { pprof.StopCPUProfile(); _ = profile.Sync(); _ = profile.Close() }) }
	defer stopProfile()
	began := time.Now()
	stop := make(chan struct{})
	defer close(stop)
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, os.Interrupt)
	defer signal.Stop(signals)
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-signals:
				stopProfile()
				fmt.Fprintln(os.Stderr, "rank interrupted; completed checkpoints retained")
				os.Exit(2)
			case <-ticker.C:
				if e := resources(c, began); e != nil {
					stopProfile()
					fmt.Fprintln(os.Stderr, e)
					os.Exit(2)
				}
			}
		}
	}()
	var random io.Reader = rand.Reader
	flags.Visit(func(f *flag.Flag) {
		if f.Name == "test-seed" {
			random = mathrand.New(mathrand.NewSource(*testSeed))
		}
	})
	if os.Getenv("LIFTRS_GPU_TAYLOR_PROGRESS") == "" {
		_ = os.Setenv("LIFTRS_GPU_TAYLOR_PROGRESS", "bounded")
	}
	if os.Getenv("LIFTRS_CPU_PROGRESS") == "" {
		_ = os.Setenv("LIFTRS_CPU_PROGRESS", "1")
	}
	// MPI calls stay on the owning thread;the watchdog only observes local resources.
	return lrs.RunDistributedPreprocessing(c, os.Stdout, random)
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
