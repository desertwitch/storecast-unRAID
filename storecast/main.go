package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/dustin/go-humanize"
	"golang.org/x/sys/unix"
)

const (
	Year  = time.Hour * 24 * 365
	Month = time.Hour * 24 * 30
	Day   = time.Hour * 24
	Hour  = time.Hour
)

type File struct {
	Size    int64
	Created time.Time
}

type Group struct {
	Size  int64
	Until time.Time
}

type Stats struct {
	CreatedMin time.Time
	CreatedMax time.Time
	Duration   time.Duration
	TotalSize  int64
	TotalCount int
}

type Forecast struct {
	Stats    Stats
	Interval time.Duration
	Slope    float64
	History  []Group
	Forecast []Group
}

type DataPoint struct {
	X string `json:"x"`
	Y int64  `json:"y"`
}

type ChartData struct {
	History   []DataPoint `json:"history"`
	Forecast  []DataPoint `json:"forecast"`
	Timestamp string      `json:"timestamp"`
}

var errorCount int32
var skippedCount int32
var normalize bool

func GenerateForecast(ctx context.Context, path string, now time.Time) (Forecast, error) {
	files, err := listFilesWithInodes(ctx, path)
	if err != nil {
		return Forecast{}, fmt.Errorf("failed to list files: %w", err)
	}

	stats, err := calculateStats(files, now)
	if err != nil {
		return Forecast{}, fmt.Errorf("failed to calculate stats: %w", err)
	}

	interval, err := deriveInterval(stats.Duration)
	if err != nil {
		return Forecast{}, fmt.Errorf("failed to derive intervals: %w", err)
	}

	history := accumulate(groupFiles(files, interval, now))
	slope := calculateRegression(history)
	forecast := estimateForecast(slope, now, interval, stats.TotalSize)

	return Forecast{
		Stats:    stats,
		Interval: interval,
		Slope:    slope,
		History:  history,
		Forecast: forecast,
	}, nil
}

func listFilesWithInodes(ctx context.Context, path string) ([]File, error) {
	var files []File

	paths := make(chan string, 1000)
	results := make(chan File, 1000)
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(paths)
		_ = filepath.WalkDir(path, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				atomic.AddInt32(&errorCount, 1)
				return nil
			}
			if !d.IsDir() {
				select {
				case paths <- path:
				case <-ctx.Done():
					return fmt.Errorf("interrupted during traversal")
				}
			}
			return nil
		})
	}()

	numWorkers := runtime.NumCPU() * 2
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range paths {
				var stat unix.Statx_t
				if err := unix.Statx(unix.AT_FDCWD, path, unix.AT_SYMLINK_NOFOLLOW, unix.STATX_BASIC_STATS|unix.STATX_BTIME, &stat); err != nil {
					atomic.AddInt32(&errorCount, 1)
					continue
				}

				birthTime := time.Unix(int64(stat.Btime.Sec), int64(stat.Btime.Nsec))
				changeTime := time.Unix(int64(stat.Ctime.Sec), int64(stat.Ctime.Nsec))

				validBirthTime := stat.Mask&unix.STATX_BTIME != 0 && birthTime.Year() >= 2000
				validChangeTime := changeTime.Year() >= 2000

				var selectedTime time.Time
				if validBirthTime && validChangeTime {
					if birthTime.Before(changeTime) {
						selectedTime = birthTime
					} else {
						selectedTime = changeTime
					}
				} else if validBirthTime {
					selectedTime = birthTime
				} else if validChangeTime {
					selectedTime = changeTime
				} else {
					atomic.AddInt32(&skippedCount, 1)
					continue
				}

				select {
				case results <- File{Size: int64(stat.Size), Created: selectedTime}:
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for file := range results {
		files = append(files, file)
	}

	if ctx.Err() != nil {
		return nil, fmt.Errorf("canceled: %w", ctx.Err())
	}

	const maxErrorThreshold = 0.5
	if atomic.LoadInt32(&errorCount) > int32(float64(len(files))*maxErrorThreshold) {
		return nil, fmt.Errorf("too many errors (>50%%) encountered during file processing")
	}

	return files, nil
}

func calculateStats(files []File, now time.Time) (Stats, error) {
	if len(files) == 0 {
		return Stats{}, fmt.Errorf("no files found")
	}
	minTime := now
	var totalSize int64
	for _, file := range files {
		if file.Created.Before(minTime) {
			minTime = file.Created
		}
		totalSize += file.Size
	}
	return Stats{
		CreatedMin: minTime,
		CreatedMax: now,
		Duration:   now.Sub(minTime),
		TotalSize:  totalSize,
		TotalCount: len(files),
	}, nil
}

func deriveInterval(duration time.Duration) (time.Duration, error) {
	switch {
	case duration >= Month*3:
		return Month, nil
	case duration >= Day*3:
		return Day, nil
	case duration >= Hour*3:
		return Hour, nil
	default:
		return 0, fmt.Errorf("all files are less than 3 hours old")
	}
}

func accumulate(groups []Group) []Group {
	var accSize int64
	for i := range groups {
		accSize += groups[i].Size
		groups[i].Size = accSize
	}
	return groups
}

func convertToDataPoints(groups []Group) []DataPoint {
	var dataPoints []DataPoint
	for _, group := range groups {
		dataPoints = append(dataPoints, DataPoint{
			X: group.Until.Format("2006-01-02 15:04"),
			Y: group.Size,
		})
	}
	return dataPoints
}

func groupFiles(files []File, interval time.Duration, now time.Time) []Group {
	groupMap := make(map[time.Time]int64)
	for _, file := range files {
		intervalKey := now.Truncate(interval).Add(-interval * time.Duration(now.Sub(file.Created)/interval))
		groupMap[intervalKey] += file.Size
	}

	var groups []Group
	for key, size := range groupMap {
		groups = append(groups, Group{Size: size, Until: key})
	}
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Until.Before(groups[j].Until)
	})
	return groups
}

func calculateRegression(history []Group) float64 {
	if len(history) < 2 {
		return 0.0
	}
	var sumX, sumY, sumXY, sumX2 float64
	for _, group := range history {
		t := float64(group.Until.Unix())
		s := float64(group.Size)
		sumX += t
		sumY += s
		sumXY += t * s
		sumX2 += t * t
	}
	meanX := sumX / float64(len(history))
	meanY := sumY / float64(len(history))

	numerator := sumXY - sumX*meanY
	denominator := sumX2 - sumX*meanX

	if denominator == 0 {
		return 0
	}
	return numerator / denominator
}

func estimateForecast(slope float64, now time.Time, dynamicInterval time.Duration, currentSize int64) []Group {
	if normalize {
		intervalRatio := float64(dynamicInterval) / float64(Month)
		normalizedSlope := slope * intervalRatio

		var forecast []Group
		for i := 1; i <= 36; i++ {
			currentSize += int64(normalizedSlope * float64(Month.Seconds()))
			forecast = append(forecast, Group{
				Size:  currentSize,
				Until: now.Add(time.Duration(i) * Month),
			})
		}
		return forecast
	} else {
		var forecast []Group
		for i := 1; i <= 36; i++ {
			currentSize += int64(slope * float64(dynamicInterval.Seconds()))
			forecast = append(forecast, Group{
				Size:  currentSize,
				Until: now.Add(time.Duration(i) * dynamicInterval),
			})
		}
		return forecast
	}
}

func formatInterval(interval time.Duration) string {
	switch interval {
	case Year:
		return "Year"
	case Month:
		return "Month"
	case Day:
		return "Day"
	case Hour:
		return "Hour"
	default:
		return interval.String() // Fallback to default formatting
	}
}

func PrintForecast(f Forecast) {
	intervalRatio := float64(f.Interval) / float64(Month)
	normalizedSlope := f.Slope * intervalRatio

	fmt.Printf("\n== Summary ==\n")
	fmt.Printf("%-20s %d\n", "Number of files:", f.Stats.TotalCount)
	fmt.Printf("%-20s %s\n", "Size of files:", humanize.Bytes(uint64(f.Stats.TotalSize)))
	fmt.Printf("%-20s %s\n", "Earliest time:", f.Stats.CreatedMin.Format("2006-01-02 15:04"))
	fmt.Printf("%-20s %s\n", "Most recent time:", f.Stats.CreatedMax.Format("2006-01-02 15:04"))
	fmt.Printf("%-20s %s\n", "Spanning:", f.Stats.Duration)
	fmt.Printf("%-20s %s\n", "Grouping:", f.Interval)
	fmt.Printf("%-20s %s / %s\n", "Slope:", humanize.Bytes(uint64(f.Slope*f.Interval.Seconds())), formatInterval(f.Interval))
	if normalize {
		fmt.Printf("%-20s %s / %s\n", "Normalized:", humanize.Bytes(uint64(normalizedSlope*Month.Seconds())), formatInterval(Month))
	}
	fmt.Printf("%-20s %d skipped - %d errors\n", "Issues:", atomic.LoadInt32(&skippedCount), atomic.LoadInt32(&errorCount))

	if (f.Slope * f.Interval.Seconds()) < 1e+9 {
		fmt.Printf("\n%-20s If the slope is nearing zero, forecasts can become more unreliable.\n", "Warning:")
	}

	fmt.Println("\n== History ==")
	for _, group := range f.History {
		fmt.Printf("%-20s %s\n", group.Until.Format("2006-01-02 15:04"), humanize.Bytes(uint64(group.Size)))
	}

	fmt.Println("\n== Forecast ==")
	for _, group := range f.Forecast {
		fmt.Printf("%-20s %s\n", group.Until.Format("2006-01-02 15:04"), humanize.Bytes(uint64(group.Size)))
	}
}

func WriteJSON(f Forecast, filePath string, pastPoints int, futurePoints int) error {
	data := ChartData{
		History:   selectSubset(f.History, pastPoints),
		Forecast:  selectSubsetFuture(f.Forecast, futurePoints),
		Timestamp: time.Now().Format("2006-01-02 15:04"),
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create JSON file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to write JSON data: %w", err)
	}

	return nil
}

func selectSubset(groups []Group, count int) []DataPoint {
	if len(groups) <= count {
		return convertToDataPoints(groups)
	}

	selected := groups[len(groups)-count:]
	return convertToDataPoints(selected)
}

func selectSubsetFuture(groups []Group, count int) []DataPoint {
	if len(groups) < 2 || count <= 1 {
		return convertToDataPoints(groups)
	}

	first := groups[0]
	last := groups[len(groups)-1]

	if count == 2 {
		return convertToDataPoints([]Group{first, last})
	}

	step := float64(len(groups)-1) / float64(count-1)
	var selected []Group

	selected = append(selected, first)

	for i := 1; i < count-1; i++ {
		index := int(step * float64(i))
		if index >= len(groups) {
			index = len(groups) - 1
		}
		selected = append(selected, groups[index])
	}

	selected = append(selected, last)

	return convertToDataPoints(selected)
}

func main() {
	fmt.Printf("Rysz's Storecast v0.1.2 (static-x86_64)\n\n")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-signalChan
		fmt.Fprintf(os.Stderr, "\n%-20s %s - exiting...\n", "Signal:", sig)
		cancel()
	}()

	path := flag.String("path", ".", "Path to the directory to analyze (required)")
	writeJSON := flag.Bool("json", false, "Write output also to JSON files for frontend")
	flag.BoolVar(&normalize, "normalize", false, "Normalize forecast to 36 months (regardless of grouping)")

	flag.Parse()

	if len(flag.Args()) > 0 {
		*path = flag.Args()[0]
	}

	fmt.Printf("%-20s %s\n", "Started:", time.Now().Format("2006-01-02 15:04"))
	fmt.Printf("%-20s %s\n", "Analyzing:", *path)
	fmt.Printf("%-20s %v\n", "Normalize:", normalize)
	fmt.Printf("%-20s %v\n\n", "Write JSON:", *writeJSON)

	fmt.Printf("Generating forecast (this may take up to 30 minutes)...\n")

	info, err := os.Stat(*path)
	if os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "%-20s Path does not exist.\n", "Error:")
		os.Exit(1)
	}

	if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "%-20s Path is not a directory.\n", "Error:")
		os.Exit(1)
	}

	forecast, err := GenerateForecast(ctx, *path, time.Now())
	if err != nil {
		fmt.Fprintf(os.Stderr, "%-20s %v\n", "Error:", err)
		os.Exit(1)
	}

	PrintForecast(forecast)

	if *writeJSON {
		var writeErrors []error

		if err := WriteJSON(forecast, "/tmp/storecast.json", 10, 10); err != nil {
			writeErrors = append(writeErrors, fmt.Errorf("storecast.json: %w", err))
		}
		if err := WriteJSON(forecast, "/tmp/storecast-dash.json", 5, 5); err != nil {
			writeErrors = append(writeErrors, fmt.Errorf("storecast-dash.json: %w", err))
		}

		if len(writeErrors) > 0 {
			for _, writeErr := range writeErrors {
				fmt.Fprintf(os.Stderr, "%-20s %v\n", "Error (JSON):", writeErr)
			}
		}
	}

	fmt.Printf("\n%-20s %s\n", "Finished:", time.Now().Format("2006-01-02 15:04"))
	os.Exit(0)
}
