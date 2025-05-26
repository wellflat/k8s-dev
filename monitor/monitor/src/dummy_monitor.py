#!/usr/bin/python3
#
# Gather System and GPU data and store in InfluxDB
#
# Requirements:
#  pip install psutil influxdb pandas matplotlib
#
# Author: Jason Cox (modified for Mac dummy GPU metrics)
# 7 May 2024
#

import subprocess
import psutil
from influxdb import InfluxDBClient
import time
import os
import sys
import signal
import pandas as pd
import matplotlib.pyplot as plt
import random  # 追加: ランダム値生成用

BUILD = "0.1"

# Replace these with your InfluxDB server details from environment variables or secrets
host = os.getenv('INFLUXDB_HOST') or 'localhost'
port = int(os.getenv('INFLUXDB_PORT')) or 8086
database = os.getenv('INFLUXDB_DATABASE') or 'llmbench'
wait_time = int(os.getenv('WAIT_TIME')) or 5

# Print application header
print(f"System and GPU Monitor v{BUILD}", file=sys.stderr)
sys.stderr.flush()

# Signal handler - Exit on SIGTERM
def sigTermHandler(signum, frame):
    raise SystemExit
signal.signal(signal.SIGTERM, sigTermHandler)

# Connect
client = InfluxDBClient(
    host=host,
    port=port,
    database=database)
# Check connection
if not client:
    print(f" - Connection to InfluxDB {host}:{port} database {database} failed", file=sys.stderr)
    sys.stderr.flush()
    sys.exit(1)
else:
    print(f" - Connection to InfluxDB {host}:{port} database {database} successful", file=sys.stderr)
    sys.stderr.flush()

# Function to run a command and return the output
def getcommand(command):
    try:
        output = subprocess.check_output(command, shell=True, universal_newlines=True)
        return output.strip()
    except subprocess.CalledProcessError as e:
        print("Error executing the command:", e)
        return ""

print(f" - Monitor started - Looping every {wait_time} seconds.", file=sys.stderr)
sys.stderr.flush()

history = []

# Main loop
try:
    while True:
        # Get system metrics
        measurements = {}
        memory_stats = psutil.virtual_memory()
        measurements["memory"] = memory_stats.used
        measurements["cpu"] = psutil.cpu_percent(interval=1.0)

        # --- GPU metrics (dummy values for Mac) ---
        # nvidia-smiはmacOSでは使えないため、ランダムなダミー値を生成
        num_gpus = 9  # シミュレーションするGPUの台数（必要に応じて変更可能）
        for i in range(num_gpus):
            # それぞれランダムな値を生成
            util = random.randint(0, 100)              # GPU利用率 [%]
            temp = round(random.uniform(30.0, 90.0), 1)  # GPU温度 [°C]
            power = round(random.uniform(20.0, 150.0), 2)  # GPU消費電力 [W]
            used = random.randint(0, 16000)              # 使用済みメモリ [MiB]
            total = 16384                              # 総GPUメモリ [MiB]（固定値の例）
            measurements[f"gpuutil{i}"] = util
            measurements[f"gputemp{i}"] = temp
            measurements[f"gpupower{i}"] = power
            measurements[f"gpumemory{i}"] = used
            measurements[f"gputotalmemory{i}"] = total

        # 保存用のタイムスタンプを取得
        timestamp = pd.Timestamp.now()
        record = {"timestamp": timestamp}
        record.update(measurements)
        history.append(record)

        # Create payload for InfluxDB
        json_body = []
        for name, value in measurements.items():
            data_point = {
                "measurement": name,
                "tags": {},
                "fields": {"value": value}
            }
            json_body.append(data_point)

        # Send to InfluxDB
        print("### json body ###", file=sys.stderr)
        print(json_body)
        print(" - Sending data to InfluxDB...", file=sys.stderr)
        print("/n")
        r = client.write_points(json_body)
        client.close()

        # Wait
        time.sleep(wait_time)

except (KeyboardInterrupt, SystemExit):
    print(" - Monitor stopped by user", file=sys.stderr)
    sys.stderr.flush()
except Exception as e:
    print(f" - Monitor stopped with error: {e}", file=sys.stderr)
    sys.stderr.flush()
finally:
    print(" - Saving collected data to CSV and PNG...", file=sys.stderr)
    sys.stderr.flush()

    if history:
        df = pd.DataFrame(history)
        df.to_csv("gpu_metrics.csv", index=False)

        # グラフ作成
        plt.figure(figsize=(12, 6))
        for col in df.columns:
            if col.startswith("gpuutil") or col.startswith("gpupower") or col.startswith("gputemp"):
                plt.plot(df["timestamp"], df[col], label=col)

        plt.legend()
        plt.xlabel("Time")
        plt.ylabel("Metric Value")
        plt.title("GPU Usage Metrics (Dummy)")
        plt.xticks(rotation=45)
        plt.tight_layout()
        plt.savefig("gpu_metrics.png")
        plt.close()

print(" - Monitor stopped", file=sys.stderr)
sys.stderr.flush()
