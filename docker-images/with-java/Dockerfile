FROM thiekus/gargoyle-judge
LABEL maintainer Thiekus

ENV USER gargoyle
ENV HOME /home/$USER
ENV GOPATH $HOME/golib

USER root

# Setting up gcc toolchain as needed by SQLite
RUN apt-get -qq update && apt-get --no-install-recommends -qq -y install gcc g++ default-jdk

# Expose master and slave ports
EXPOSE 28498
EXPOSE 28499

# Run our server, done!
CMD [ "/opt/gargoyle/bin/gyslave" ]
