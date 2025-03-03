FROM registry.cn-shenzhen.aliyuncs.com/starsl/python:3.11-alpine
ADD . /
WORKDIR /
RUN pip3 install -r requirements.txt -i https://pypi.tuna.tsinghua.edu.cn/simple/ --no-cache-dir
EXPOSE 80
CMD ["python3", "/kubedoor-master.py"]
