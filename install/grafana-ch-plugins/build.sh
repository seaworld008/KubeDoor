#!/bin/bash
docker build -t grafana-plugins-init:0.1.0 .
docker login --username=starsliao@163.com registry.cn-shenzhen.aliyuncs.com
docker tag grafana-plugins-init:0.1.0 registry.cn-shenzhen.aliyuncs.com/starsl/grafana-plugins-init:0.1.0
docker push registry.cn-shenzhen.aliyuncs.com/starsl/grafana-plugins-init:0.1.0