FROM centos:7
MAINTAINER poppinzhang zhangyu960310@gmail.com
WORKDIR /root
USER root
RUN set -eux; \
    yum install -y wget tar bzip2 gzip; \
    wget https://github.com/zhangyu0310/wechat-setu-bot/releases/download/v0.2.0/setu_server.tgz; \
    tar xzf setu_server.tgz;

CMD while true;do cd /root/wechat-setu-bot;./start_up.sh;sleep 30; done
