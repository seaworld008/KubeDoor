#!/bin/bash
PASSWORD=$(base64 < /dev/urandom | head -c8)
mkdir -p /opt/clickhouse/etc/clickhouse-server/{config.d,users.d}
cd /opt/clickhouse

cat <<-EOF > etc/kubedoor-init.sql
CREATE DATABASE IF NOT EXISTS kubedoor ENGINE=Atomic;

CREATE TABLE IF NOT EXISTS kubedoor.k8s_agent_status
(
    env String,
    collect Bool DEFAULT false,
    peak_hours String,
    admission Bool DEFAULT false,
    admission_namespace String,
    nms_not_confirm Bool DEFAULT false,
    scheduler Bool DEFAULT false
)
ENGINE = MergeTree
PRIMARY KEY tuple(env)
SETTINGS index_granularity = 8192;

CREATE TABLE IF NOT EXISTS kubedoor.k8s_res_control
(
    env String,
    namespace String,
    deployment String,
    pod_count_init UInt8 DEFAULT 0 COMMENT '初始化pod数,仅首次记录该值,从当前pod数复制',
    pod_count UInt8 DEFAULT 0 COMMENT '当前pod数,第三优先级',
    pod_count_manual Int16 DEFAULT -1 COMMENT '手动填写需要的pod数,第一优先级,webhook拦截时读取到-1,则取pod_count_ai',
    p95_pod_cpu_pct Float32 DEFAULT -1,
    p95_pod_mem_pct Float32 DEFAULT -1,
    request_cpu_m Int32 DEFAULT -1 COMMENT '取自高峰期的p95 CPU负载,最小1,新建服务时,该值默认为-1,webhook拦截时读取到-1,则不改变改字段,并通知。',
    request_mem_mb Int32 DEFAULT -1 COMMENT '取自高峰期的p95 内存使用量,最小1,新建服务时,该值默认为-1,webhook拦截时读取到-1,则不改变改字段,并通知。',
    limit_cpu_m Int32 DEFAULT -1 COMMENT '初始化的时候取当前pod的limit cpu,没有则为-1,webhook拦截时读取到-1,则不改变改字段,并通知。',
    limit_mem_mb Int32 DEFAULT -1 COMMENT '初始化的时候取当前pod的limit mem,没有则为-1,webhook拦截时读取到-1,则不改变改字段,并通知。',
    update DateTime('Asia/Shanghai'),
    pod_mem_saved_mb Float32 DEFAULT -1,
    pod_qps Float32 DEFAULT -1,
    pod_g1gc_qps Float32 DEFAULT -1,
    pod_count_ai Int16 DEFAULT -1 COMMENT 'AI计算后需要的pod数,第二优先级,webhook拦截时读取到-1,则取pod_count',
    pod_qps_ai Float32 DEFAULT -1,
    pod_load_ai Float32 DEFAULT -1,
    pod_g1gc_qps_ai Float32 DEFAULT -1,
    update_ai DateTime('Asia/Shanghai')
)
ENGINE = MergeTree
PRIMARY KEY (env,namespace,deployment)
ORDER BY (env,namespace,deployment)
SETTINGS index_granularity = 8192;


CREATE TABLE IF NOT EXISTS kubedoor.k8s_resources
(
    date DateTime('Asia/Shanghai'),
    env String,
    namespace String,
    deployment String,
    pod_count UInt8,
    p95_pod_load Float32,
    p95_pod_cpu_pct Float32,
    p95_pod_wss_mb Float32,
    p95_pod_wss_pct Float32,
    limit_pod_cpu_m Float32,
    limit_pod_mem_mb Float32,
    request_pod_cpu_m Float32,
    request_pod_mem_mb Float32,
    p95_pod_qps Float32,
    p95_pod_g1gc_qps Float32,
    pod_jvm_max_mb Float32
)
ENGINE = MergeTree
PARTITION BY toYYYYMM(date)
PRIMARY KEY (date,env,namespace,deployment)
ORDER BY (date,env,namespace,deployment)
TTL toDateTime(date) + toIntervalDay(365)
SETTINGS index_granularity = 8192;

CREATE TABLE IF NOT EXISTS kubedoor.k8s_pod_alert_days
(
    fingerprint String,
    alert_status String,
    send_resolved Bool DEFAULT true,
    count_firing UInt32,
    count_resolved Int32,
    start_time DateTime('Asia/Shanghai'),
    end_time Nullable(DateTime('Asia/Shanghai')) DEFAULT NULL,
    severity String,
    alert_group String,
    alert_name String,
    env String,
    namespace String,
    container String,
    pod String,
    description String,
    operate String
)
ENGINE = MergeTree
PARTITION BY toYYYYMMDD(start_time)
PRIMARY KEY (start_time,
  fingerprint,
  severity,
  env,
  alert_group,
  alert_name)
TTL toDateTime(start_time) + toIntervalDay(365)
SETTINGS index_granularity = 8192;
EOF

cat <<-EOF > docker-compose.yaml
services:
  clickhouse:
    image: registry.cn-shenzhen.aliyuncs.com/starsl/clickhouse-server:24.8-alpine
    container_name: clickhouse
    hostname: clickhouse
    volumes:
      - /opt/clickhouse/logs:/var/log/clickhouse-server
      - /opt/clickhouse/data:/var/lib/clickhouse
      - /opt/clickhouse/etc/kubedoor-init.sql:/docker-entrypoint-initdb.d/kubedoor-init.sql
      - /opt/clickhouse/etc/clickhouse-server/config.d/config.xml:/etc/clickhouse-server/config.d/config.xml
      - /opt/clickhouse/etc/clickhouse-server/users.d/users.xml:/etc/clickhouse-server/users.d/users.xml

    restart: always
    environment:
      - TZ=Asia/Shanghai
      - CLICKHOUSE_DB=kubedoor
      - CLICKHOUSE_USER=default
      - CLICKHOUSE_PASSWORD=$PASSWORD
      - CLICKHOUSE_DEFAULT_ACCESS_MANAGEMENT=1
    ulimits:
      nofile:
        soft: 262144
        hard: 262144
    ports:
      - 8123:8123
      - 9000:9000
    network_mode: host
EOF

cat <<-EOF > etc/clickhouse-server/config.d/config.xml
<clickhouse replace="true">
    <logger>
        <level>warning</level>
        <log>/var/log/clickhouse-server/clickhouse-server.log</log>
        <errorlog>/var/log/clickhouse-server/clickhouse-server.err.log</errorlog>
        <size>1000M</size>
        <count>3</count>
    </logger>
    <display_name>my_clickhouse</display_name>
    <listen_host>0.0.0.0</listen_host>
    <http_port>8123</http_port>
    <tcp_port>9000</tcp_port>
    <user_directories>
        <users_xml>
            <path>users.xml</path>
        </users_xml>
        <local_directory>
            <path>/var/lib/clickhouse/access/</path>
        </local_directory>
    </user_directories>
</clickhouse>
EOF

encrypt=`echo -n "$PASSWORD" | sha256sum | tr -d ' |-'`

cat <<-EOF > etc/clickhouse-server/users.d/users.xml
<?xml version="1.0"?>
<clickhouse replace="true">
    <profiles>
        <default>
            <local_filesystem_read_method>pread</local_filesystem_read_method>
            <max_memory_usage>10000000000</max_memory_usage>
            <use_uncompressed_cache>0</use_uncompressed_cache>
            <load_balancing>in_order</load_balancing>
            <log_queries>1</log_queries>
        </default>
    </profiles>
    <users>
        <default>
            <password remove='1' />
            <password_sha256_hex>$encrypt</password_sha256_hex>
            <access_management>1</access_management>
            <profile>default</profile>
            <networks>
                <ip>::/0</ip>
            </networks>
            <quota>default</quota>
            <access_management>1</access_management>
            <named_collection_control>1</named_collection_control>
            <show_named_collections>1</show_named_collections>
            <show_named_collections_secrets>1</show_named_collections_secrets>
        </default>
    </users>
    <quotas>
        <default>
            <interval>
                <duration>3600</duration>
                <queries>0</queries>
                <errors>0</errors>
                <result_rows>0</result_rows>
                <read_rows>0</read_rows>
                <execution_time>0</execution_time>
            </interval>
        </default>
    </quotas>
</clickhouse>
EOF
echo -e "ClickHouse 部署完成\n默认用户: default\n密码: $PASSWORD\n请执行以下命令进入目录并启动ClickHouse:\n\ncd /opt/clickhouse && docker compose up -d\n"
