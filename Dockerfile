FROM golang:1.14-buster as builder

ENV GOPATH=/root/go
RUN mkdir -p /app
COPY ./ /app
RUN cd /app \
    && make build

FROM debian:buster-slim
MAINTAINER David Prandzioch <vaskovasilev94@yahoo.com>

RUN DEBIAN_FRONTEND=noninteractive apt-get update && \
	apt-get install -q -y bind9 dnsutils && \
	apt-get clean

RUN chmod 770 /var/cache/bind
COPY setup.sh /root/setup.sh
RUN chmod +x /root/setup.sh
COPY named.conf.options /etc/bind/named.conf.options
COPY --from=builder /app/ddns /root/ddns

EXPOSE 53 8080
CMD ["sh", "-c", "/root/setup.sh ; service bind9 start ; /root/ddns"]
