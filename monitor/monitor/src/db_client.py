#!/usr/bin/python3
"""
Retrieve GPU utilization metrics from InfluxDB and plot them.

Requirements:
  pip install influxdb pandas matplotlib

※ このコードは、先にGPUメトリクスがInfluxDBに書き込まれていることを前提としています。
    測定値のmeasurement名は "gpuutil0", "gpuutil1", ... と仮定しています。

Author: Your Name
Date: 2024-05-XX
"""

import os
import sys
import pandas as pd
import matplotlib.pyplot as plt
from influxdb import InfluxDBClient

# InfluxDBの接続情報（環境変数があればそちらを利用）
host = os.getenv('INFLUXDB_HOST') or 'localhost'
port = int(os.getenv('INFLUXDB_PORT') or 8086)
database = os.getenv('INFLUXDB_DATABASE') or 'llmbench'

print(f"Connecting to InfluxDB at {host}:{port}, Database: {database}", file=sys.stderr)
client = InfluxDBClient(host=host, port=port, database=database)

# GPUの利用率情報は、書き込み時に "gpuutil0", "gpuutil1", … というmeasurement名で保存されているとする
# 正規表現を使用して "gpuutil" を含むすべてのmeasurementからデータを取得します
# query = "SELECT * FROM /gpuutil/"  # measurement名に "gpuutil" を含む全てのデータを取得
# query = 'SELECT max(\"value\") FROM \"gpupower6\" WHERE $timeFilter GROUP BY time($__interval) fill(null)'
query = '''
SELECT max("value") 
FROM "gpupower0" 
WHERE time >= '2025-04-15T07:00:00.223702Z' 
  AND time < '2025-04-15T09:00:00.223702Z'
GROUP BY time(5m) fill(null)
'''
print(f"Executing query: {query}", file=sys.stderr)
result = client.query(query)

print(result, file=sys.stderr)

# # InfluxDBから取得した結果（ResultSet）は複数のseries（各measurementごと）となるため、各seriesをDataFrameに変換
# dfs = {}
# for series, points in result.items():
#     # series はタプル( measurement_name, ... ) となっているので最初の要素を利用
#     measurement_name = series[0]
#     data = list(points)
#     if data:
#         df = pd.DataFrame(data)
#         # InfluxDBのtimeは文字列なので、日付型に変換
#         df['time'] = pd.to_datetime(df['time'])
#         dfs[measurement_name] = df
#         print(f"Retrieved {len(df)} points for {measurement_name}", file=sys.stderr)
#     else:
#         print(f"No data retrieved for {measurement_name}", file=sys.stderr)

# # 取得した各GPU利用率のDataFrameをそれぞれCSVに出力（オプション）
# for measurement_name, df in dfs.items():
#     csv_filename = f"{measurement_name}.csv"
#     df.to_csv(csv_filename, index=False)
#     print(f"Saved {measurement_name} data to {csv_filename}", file=sys.stderr)

# # 各measurementごとに時系列プロットを作成する（例：GPU利用率は%であると仮定）
# plt.figure(figsize=(12, 6))
# for measurement_name, df in dfs.items():
#     # 'value'フィールドに利用率の数値が保存されていると仮定
#     plt.plot(df['time'], df['value'], label=measurement_name)
# plt.xlabel("Time")
# plt.ylabel("GPU Utilization (%)")
# plt.title("GPU Utilization Metrics from InfluxDB")
# plt.legend()
# plt.xticks(rotation=45)
# plt.tight_layout()
# plt.savefig("retrieved_gpu_utilization.png")
# plt.show()

client.close()
print("InfluxDB connection closed.", file=sys.stderr)
