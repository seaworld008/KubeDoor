FROM registry.cn-shenzhen.aliyuncs.com/starsl/busybox:1.37.0
RUN mkdir -p /app/plugins
COPY plugins /app/plugins
WORKDIR /app
CMD ["sh", "-c", "cp -r ./plugins/ /var/lib/grafana/plugins/"]
