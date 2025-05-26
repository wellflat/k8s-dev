FROM nvidia/cuda:12.8.1-devel-ubuntu22.04
RUN apt-get update && apt-get install -y python3 python3-pip
RUN pip3 install psutil influxdb pandas matplotlib
WORKDIR /workspace
COPY ./src/monitor.py .
RUN mkdir metrics
CMD ["python3", "monitor.py"]
