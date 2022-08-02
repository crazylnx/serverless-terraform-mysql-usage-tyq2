FROM golang:1.18.1 AS builder


WORKDIR /workspace

COPY go.mod go.mod
COPY go.sum go.sum

COPY code/ code/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o server ./code/server.go

FROM registry.cn-huhehaote.aliyuncs.com/fc-demo2/terraform-binary:0.0.1 as binary

FROM alpine
WORKDIR /

COPY --from=builder /workspace/server .
COPY --from=binary bin/.terraform.d /root/.terraform.d
COPY --from=binary bin/terraform /usr/bin/terraform

COPY data/data.tf /data/data.tf
COPY data/rds     /data/rds


ENTRYPOINT ["/server"]
