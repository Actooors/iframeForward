FROM ubuntu
MAINTAINER mzz2017 m@mzz.pub
WORKDIR /iframeForward/
ADD iframeForward ./
ADD sources.list /etc/apt/sources.list
RUN apt-get update
RUN apt-get install -y ca-certificates
EXPOSE 8090
#ENV GIN_MODE release
ENTRYPOINT ["./iframeForward"]
