#!/bin/bash

# 证书请求文件配置
expiredays=36500  #制定证书过期时间100年
service="kubedoor-agent"
namespace="kubedoor"
svc="svc"
cluster="cluster.local"

mkdir -p ssl
cd ssl

cat > csr.conf <<EOF
[req]
req_extensions = v3_req
distinguished_name = req_distinguished_name
default_bits = 2048
prompt = no
default_md = sha256

[v3_req]

[req_distinguished_name]
C = CN
ST = SZ
L = SZ
O = StarsL
OU = StarsL
CN = $service.$namespace

[ v3_ext ]
authorityKeyIdentifier=keyid,issuer:always
basicConstraints=CA:FALSE
keyUsage=keyEncipherment,dataEncipherment
extendedKeyUsage=serverAuth,clientAuth
subjectAltName=@alt_names

[alt_names]
DNS.1 = $service
DNS.2 = $service.$namespace
DNS.3 = $service.$namespace.$svc
DNS.4 = $service.$namespace.$svc.$cluster

EOF

# 创建证书
openssl genrsa -out ca.key 2048
openssl req -x509 -new -nodes -key ca.key -subj "/CN=$service.$namespace.$svc" -days $expiredays -out ca.crt

openssl genrsa -out tls.key 2048

openssl req -new -key tls.key -out tls.csr -config csr.conf

openssl x509 -req -in tls.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out tls.crt -days $expiredays -extensions v3_ext -extfile csr.conf

cd ..

#替换CA
base64ca=`cat ssl/ca.crt|base64`
#sed -i -e "s/caBundle: .*/caBundle: $base64ca/g" mutatingwebhookconfiguration.yaml
echo $base64ca