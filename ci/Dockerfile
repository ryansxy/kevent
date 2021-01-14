FROM ubuntu:16.04

MAINTAINER shaxiangyu@baidu.com

RUN apt-get update && apt-get install curl apt-transport-https --yes

RUN curl https://baltocdn.com/helm/signing.asc | apt-key add - && echo "deb https://baltocdn.com/helm/stable/debian/ all main" | tee /etc/apt/sources.list.d/helm-stable-debian.list && apt-get update && apt-get install helm

ADD kevent /usr/bin/kevent

CMD ["/usr/bin/kevent"]