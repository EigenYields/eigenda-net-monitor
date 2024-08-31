package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

var (
	batchLatency = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "damon_last_batch_latency",
			Help: "Time the last batch took to download in seconds.",
		},
		[]string{"interface"},
	)

	overallBatchLatency = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "damon_average_batch_latency",
			Help: "Average time taken for all batches to download.",
		},
		[]string{"interface"},
	)

	batchTransferredMiB = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "damon_total_transferred_MiB",
			Help: "Total data transferred in MiB for the current batch.",
		},
		[]string{"interface"},
	)
	batchAverageSpeedMiBps = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "damon_average_speed_MiBps",
			Help: "Average download speed in MiB/s for the current batch.",
		},
		[]string{"interface"},
	)
	overallAverageSpeedMiBps = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "damon_overall_average_speed_MiBps",
			Help: "Overall average download speed in MiB/s across all batches.",
		},
		[]string{"interface"},
	)
	overallAverageSpeedMBps = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "damon_overall_average_speed_MBps",
			Help: "Overall average download speed in MB/s across all batches.",
		},
		[]string{"interface"},
	)
	batchesObserved = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "damon_batches_observed_total",
			Help: "Total number of download batches observed.",
		},
		[]string{"interface"},
	)
	totalDataTransferred = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "damon_overall_data_transferred_MiB",
			Help: "Total amount of data transferred across all batches in MiB.",
		},
		[]string{"interface"},
	)
)

func init() {
	prometheus.MustRegister(batchTransferredMiB)
	prometheus.MustRegister(batchAverageSpeedMiBps)
	prometheus.MustRegister(overallAverageSpeedMiBps)
	prometheus.MustRegister(overallAverageSpeedMBps)
	prometheus.MustRegister(batchesObserved)
	prometheus.MustRegister(totalDataTransferred)
	prometheus.MustRegister(batchLatency)
	prometheus.MustRegister(overallBatchLatency)

	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logrus.SetLevel(logrus.InfoLevel)
}

func main() {

	interfaceName := flag.String("interface", "ens0", "The network interface to monitor")
	promPort := flag.String("port", "2112", "Prometheus metrics server port")
	debug := flag.Bool("debug", false, "Enable debug logging")
	softNet := flag.Bool("softnet", false, "Log softnet stats")

	flag.Parse()

	if *debug {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	// Configurable options
	interval := time.Millisecond * 250
	var minDownloadThreshold = int(1048576 * interval.Seconds()) // 1 MiB/s
	// Initalize
	totalBytes := 0
	totalTransferredMiB := 0.0
	totalSpeedMiBps := 0.0
	totalSpeedMBps := 0.0
	batchesObservedCount := 0
	overallBatchMiBps := 0.0
	overallBatchMBps := 0.0
	overallAvgLatency := 0.0
	totalLatencyAllBatches := 0.0
	downloadStarted := false

	http.Handle("/metrics", promhttp.Handler())
	go func() {
		logrus.Info(fmt.Printf("Prometheus metrics server started on port %s", *promPort))
		logrus.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", *promPort), nil))
	}()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	rxBytesBefore := getRxBytes(*interfaceName) // Initialize before the loop
	var start time.Time

	for range ticker.C {
		rxBytesAfter := getRxBytes(*interfaceName) // Measure after the interval
		rxBytesDiff := rxBytesAfter - rxBytesBefore

		if rxBytesDiff >= minDownloadThreshold {
			if !downloadStarted {
				start = time.Now()
				downloadStarted = true
			}
			totalBytes += rxBytesDiff

			elapsed := time.Since(start).Seconds()

			speedMiBps := (float64(rxBytesDiff) / 1024 / 1024) / elapsed // Convert bytes to MiB, then divide by elapsed time
			speedMBps := (float64(rxBytesDiff) / 1000000) / elapsed      // Convert bytes to MB, then divide by elapsed time
			intervalTransferredMiB := float64(rxBytesDiff) / (1024 * 1024)

			totalSpeedMiBps += speedMiBps
			totalSpeedMBps += speedMBps

			logrus.WithFields(logrus.Fields{
				"MiB/s":      speedMiBps,
				"MB/s":       speedMBps,
				"size_MiB":   intervalTransferredMiB,
				"bytes_diff": rxBytesDiff,
			}).Debug("Rx bytes detected")
		}

		if downloadStarted && rxBytesDiff < minDownloadThreshold {
			latency := time.Since(start).Seconds()
			totalLatencyAllBatches += latency

			averageSpeedMiBpsValue := totalSpeedMiBps / float64(batchesObservedCount+1)
			averageSpeedMBpsValue := totalSpeedMBps / float64(batchesObservedCount+1)

			finalTransferredMiB := float64(totalBytes) / (1024 * 1024)
			totalTransferredMiB += finalTransferredMiB

			logrus.WithFields(logrus.Fields{
				"transferred_MiB": finalTransferredMiB,
				"latency_secs":    latency,
				"bytes":           totalBytes,
				"batch_avg_MiB/s": averageSpeedMiBpsValue,
				"batch_avg_MB/s":  averageSpeedMBpsValue,
			}).Info("Batch average")

			batchesObservedCount++
			overallBatchMiBps = (overallBatchMiBps*float64(batchesObservedCount-1) + averageSpeedMiBpsValue) / float64(batchesObservedCount)
			overallBatchMBps = (overallBatchMBps*float64(batchesObservedCount-1) + averageSpeedMBpsValue) / float64(batchesObservedCount)
			overallAvgLatency = totalLatencyAllBatches / float64(batchesObservedCount)

			logrus.WithFields(logrus.Fields{
				"overall_avg_MiB/s":     overallBatchMiBps,
				"overall_avg_MB/s":      overallBatchMBps,
				"batches_observed":      batchesObservedCount,
				"total_MiB_transferred": totalTransferredMiB,
				"average_latency_secs":  overallAvgLatency,
			}).Info("Overall Average")

			// Prometheus metrics update
			batchTransferredMiB.WithLabelValues(*interfaceName).Set(finalTransferredMiB)
			batchAverageSpeedMiBps.WithLabelValues(*interfaceName).Set(averageSpeedMiBpsValue)
			overallAverageSpeedMiBps.WithLabelValues(*interfaceName).Set(overallBatchMiBps)
			overallAverageSpeedMBps.WithLabelValues(*interfaceName).Set(overallBatchMBps)
			batchesObserved.WithLabelValues(*interfaceName).Set(float64(batchesObservedCount))
			totalDataTransferred.WithLabelValues(*interfaceName).Set(totalTransferredMiB)
			batchLatency.WithLabelValues(*interfaceName).Set(latency)
			overallBatchLatency.WithLabelValues(*interfaceName).Set(overallAvgLatency)

			// Log softnet stats if enabled
			if *softNet {
				softnetStats := getSoftnetStats()
				logrus.Debug(softnetStats)
			}

			// Reset for next batch
			downloadStarted = false
			totalBytes = 0
			totalSpeedMiBps = 0
			totalSpeedMBps = 0
		}

		rxBytesBefore = rxBytesAfter
	}
}

func getRxBytes(interfaceName string) int {
	data, err := os.ReadFile(fmt.Sprintf("/sys/class/net/%s/statistics/rx_bytes", interfaceName))
	if err != nil {
		logrus.Fatalf("Failed to read rx_bytes: %v", err)
	}
	rxBytes, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		logrus.Fatalf("Failed to convert rx_bytes to int: %v", err)
	}
	return rxBytes
}

func getSoftnetStats() string {
	data, err := os.ReadFile("/proc/net/softnet_stat")
	if err != nil {
		logrus.Fatalf("Failed to read softnet_stat: %v", err)
	}
	lines := strings.Split(string(data), "\n")
	result := ""
	for i, line := range lines {
		if len(line) > 0 {
			fields := strings.Fields(line)
			dropped := hexToDec(fields[1])
			squeezed := hexToDec(fields[2])
			collision := hexToDec(fields[3])
			rps := hexToDec(fields[4])
			ipi := hexToDec(fields[5])
			queued := hexToDec(fields[6])
			rpsFail := hexToDec(fields[7])
			result += fmt.Sprintf("CPU%d: dropped=%d squeezed=%d collision=%d rps=%d ipi=%d queued=%d rps_fail=%d\n", i, dropped, squeezed, collision, rps, ipi, queued, rpsFail)
		}
	}
	return result
}

func hexToDec(hexStr string) int {
	result, err := strconv.ParseInt(hexStr, 16, 64)
	if err != nil {
		logrus.Fatalf("Failed to convert hex to int: %v", err)
	}
	return int(result)
}
