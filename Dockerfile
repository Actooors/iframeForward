FROM ubuntu
MAINTAINER mzz2017 m@mzz.pub
WORKDIR /iframeForward/
ADD iframeForward ./
EXPOSE 8090
ENTRYPOINT ["./iframeForward"]
