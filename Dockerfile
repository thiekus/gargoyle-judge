FROM ubuntu:18.04 as base

ENV USER gargoyle-build
ENV HOME /home/$USER
ENV GOPATH $HOME/golib

RUN apt-get -qq update && apt-get -qq upgrade && apt-get --no-install-recommends -qq -y install build-essential ca-certificates curl git
RUN GO=go1.13.linux-amd64.tar.gz && curl -sL --retry 10 --retry-delay 60 -O https://dl.google.com/go/$GO && tar -xzf $GO -C /usr/local
#RUN git clone https://github.com/thiekus/gargoyle-judge.git $HOME
