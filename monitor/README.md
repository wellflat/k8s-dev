# Monitoring Dashboard

This sets up a Grafana dashboard to monitor your LLM system. A [monitor.py](https://github.com/Liquid-dev/llm-bench/blob/develop/monitoring/monitor/src/monitor.py) script to pull the GPU metrics from the nvidia-smi tool.

![image](https://github.com/jasonacox/TinyLLM/assets/836718/a1389cb0-c3d1-46ec-bec1-1ff3ac412507)
![image](https://github.com/vllm-project/vllm/assets/836718/878b4c99-2707-4907-9847-6521aad30755)

## Monitoring Tool

The monitor.py script will poll local Nvidia GPU and host CPU information and store it in the InfluxDB for graphing in Grafana. The steps below will build a CUDA container that will fetch the metrics every 5 seconds.

```bash
docker compose -f compose.monitor.yaml up
```

## Dashboard Setup

Dashboard Setup

1. Go to `http://localhost:3000` and default user/password is admin/admin.
2. Create a data source, select InfluxDB and use URL http://influxdb:8086 (replace with IP address of host), database name `llmbench` and timeout `5s`.
3. Import dashboard and upload or copy/paste [dashboard.json](dashboard.json). Select InfluxDB as the data source.
