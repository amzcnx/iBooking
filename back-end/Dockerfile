#依赖镜像
FROM golang:1.20.4

#作者信息
MAINTAINER "mtf"

# 配置模块代理
ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn,direct

#工作目录
WORKDIR /opt
ADD . /opt

#在Docker工作目录下执行命令
RUN go build -o main ./main.go

#暴露端口
EXPOSE 8800

#执行项目的命令
CMD ["/opt/main"]
