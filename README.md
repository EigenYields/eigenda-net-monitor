# Eigenda Net Monitor

`eigenda-net-monitor` is a small go app to help monitor EigenDA batch end to end latency and download speed from the disperser. 

## What it does

It simply monitors the `rx_bytes` from the network interface on your EigenDA instance, if recieved bytes are less than < 1 MiB/s, it does nothing. If the rate if received bytes increases above 1 MiB/s we assume a batch has started being downloaded from the disperser and wait for it to finish to calculate metrics. Therefore, if your instance downloads any other file it will assume it is a batch. You may need to tweak the threshold if batches are being double counted, but these setting work well for us.

We use the last batch latency in combination with the request latency StoreChunks RPC metric to get an idea of the total end to end latency.

## Prometheus Metrics

The following metrics are exposed and can be scraped by Prometheus:

- `damon_total_transferred_MiB`: Data transferred in MiB for the last batch.
- `damon_average_speed_MiBps`: Download speed in MiB/s for the last batch.
- `damon_overall_average_speed_MiBps`: Overall average download speed in MiB/s across all batches.
- `damon_overall_average_speed_MBps`: Overall average download speed in MB/s across all batches.
- `damon_batches_observed_total`: Total number of batches observed.
- `damon_total_data_transferred_MiB`: Total amount of data transferred across all batches in MiB.
- `damon_overall_average_latency_seconds`: Overall average latency in seconds across all batches.
- `damon_last_batch_latency`: Latency of the last batch in seconds.

## Installation

### Prerequisites

- Go 1.16 or later
- Prometheus for scraping metrics

### Clone the Repository

Clone the `eigenda-net-monitor` repository to your local machine:
```bash
git clone https://github.com/eigenyields/eigenda-net-monitor.git 
cd eigenda-net-monitor
```

### Build
To build run:
```bash
go mod download
go build -o build/eigenda-net-monitor
```

## Run
Specify the network interface i.e ens5
```bash
./build/eigenda-net-monitor --interface <network interface> --debug
```

To scrape using promtheus you can use the following snippet in prometheus.yml

```yaml
scrape_configs:
  - job_name: 'eigenda-net-monitor'
    static_configs:
      - targets: ['localhost:2112']
```

### Example output
```bash
DEBU[2024-08-30T12:48:57Z] Rx bytes detected                             MB/s=207.064692 MiB/s=197.4722785949707 bytes_diff=51766173 size_MiB=49.368069648742676
DEBU[2024-08-30T12:48:57Z] Rx bytes detected                             MB/s=296.651272 MiB/s=282.90869903564453 bytes_diff=74162818 size_MiB=70.72717475891113
DEBU[2024-08-30T12:48:58Z] Rx bytes detected                             MB/s=246.65914 MiB/s=235.23248672485352 bytes_diff=61664785 size_MiB=58.80812168121338
DEBU[2024-08-30T12:48:58Z] Rx bytes detected                             MB/s=239.71028 MiB/s=228.60553741455078 bytes_diff=59927570 size_MiB=57.151384353637695
DEBU[2024-08-30T12:48:58Z] Rx bytes detected                             MB/s=253.051736 MiB/s=241.32894134521484 bytes_diff=63262934 size_MiB=60.33223533630371
DEBU[2024-08-30T12:48:58Z] Rx bytes detected                             MB/s=183.220224 MiB/s=174.732421875 bytes_diff=45805056 size_MiB=43.68310546875
DEBU[2024-08-30T12:48:59Z] Rx bytes detected                             MB/s=184.716688 MiB/s=176.15956115722656 bytes_diff=46179172 size_MiB=44.03989028930664
DEBU[2024-08-30T12:48:59Z] Rx bytes detected                             MB/s=198.140804 MiB/s=188.96179580688477 bytes_diff=49535201 size_MiB=47.24044895172119
DEBU[2024-08-30T12:48:59Z] Rx bytes detected                             MB/s=191.829728 MiB/s=182.94308471679688 bytes_diff=47957432 size_MiB=45.73577117919922
DEBU[2024-08-30T12:48:59Z] Rx bytes detected                             MB/s=196.669712 MiB/s=187.55885314941406 bytes_diff=49167428 size_MiB=46.889713287353516
DEBU[2024-08-30T12:49:00Z] Rx bytes detected                             MB/s=186.916084 MiB/s=178.2570686340332 bytes_diff=46729021 size_MiB=44.5642671585083
DEBU[2024-08-30T12:49:00Z] Rx bytes detected                             MB/s=196.39882 MiB/s=187.30051040649414 bytes_diff=49099705 size_MiB=46.825127601623535
DEBU[2024-08-30T12:49:00Z] Rx bytes detected                             MB/s=190.014128 MiB/s=181.2115936279297 bytes_diff=47503532 size_MiB=45.30289840698242
DEBU[2024-08-30T12:49:00Z] Rx bytes detected                             MB/s=256.624504 MiB/s=244.73619842529297 bytes_diff=64156126 size_MiB=61.18404960632324
DEBU[2024-08-30T12:49:01Z] Rx bytes detected                             MB/s=198.596524 MiB/s=189.39640426635742 bytes_diff=49649131 size_MiB=47.349101066589355
DEBU[2024-08-30T12:49:01Z] Rx bytes detected                             MB/s=202.502228 MiB/s=193.12117385864258 bytes_diff=50625557 size_MiB=48.280293464660645
DEBU[2024-08-30T12:49:01Z] Rx bytes detected                             MB/s=194.567128 MiB/s=185.55367279052734 bytes_diff=48641782 size_MiB=46.388418197631836
DEBU[2024-08-30T12:49:01Z] Rx bytes detected                             MB/s=196.963564 MiB/s=187.83909225463867 bytes_diff=49240891 size_MiB=46.95977306365967
DEBU[2024-08-30T12:49:02Z] Rx bytes detected                             MB/s=195.598992 MiB/s=186.53773498535156 bytes_diff=48899748 size_MiB=46.63443374633789
DEBU[2024-08-30T12:49:02Z] Rx bytes detected                             MB/s=138.847388 MiB/s=132.41518783569336 bytes_diff=34711847 size_MiB=33.10379695892334
INFO[2024-08-30T12:49:02Z] Batch average                                 batch_avg_MB/s=207.73718179999997 batch_avg_MiB/s=198.1136148452759 bytes=1038685909 latency_secs=5 transferred_MiB=990.5680742263794
INFO[2024-08-30T12:49:02Z] Overall Average                               average_latency_secs=5 batches_observed=1 overall_avg_MB/s=207.73718179999997 overall_avg_MiB/s=198.1136148452759 total_MiB_transferred=990.5680742263794
```
