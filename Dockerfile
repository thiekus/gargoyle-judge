FROM golang:1.13.1-stretch as base

ENV USER gargoyle
ENV HOME /home/$USER
ENV GOPATH $HOME/golib

# Setting up gcc toolchain as needed by SQLite
RUN apt-get -qq update && apt-get --no-install-recommends -qq -y install gcc

# Create new non-root user
RUN useradd -ms /bin/bash $USER

# Hit the new user
USER $USER
RUN cd $HOME && mkdir $GOPATH && mkdir bin && mkdir lib && mkdir lib/caches && mkdir lib/cert && mkdir lib/logs

# Copy required build files
COPY --chown=$USER:$USER build.sh $HOME/build.sh
COPY --chown=$USER:$USER go.mod $HOME/go.mod
COPY --chown=$USER:$USER go.sum $HOME/go.sum
COPY --chown=$USER:$USER gymaster $HOME/gymaster
COPY --chown=$USER:$USER gyslave $HOME/gyslave
COPY --chown=$USER:$USER internal $HOME/internal
COPY --chown=$USER:$USER lib/assets $HOME/lib/assets
COPY --chown=$USER:$USER lib/templates $HOME/lib/templates
COPY --chown=$USER:$USER lib/default.sql $HOME/lib/default.sql
COPY --chown=$USER:$USER lib/favicon.ico $HOME/lib/favicon.ico

# Build our server
RUN cd $HOME && sh ./build.sh

# Now create the barebone image
FROM ubuntu:18.04
LABEL maintainer Thiekus

ENV USER gargoyle
ENV HOME /home/$USER
ENV INSTALLDIR /opt/gargoyle

# Basic C/C++ and Java OpenJDK toolchains
RUN apt-get -qq update && apt-get -qq upgrade && apt-get --no-install-recommends -qq -y install sudo

# Create new non-root user, but add as sudo
RUN useradd -ms /bin/bash $USER && usermod -aG sudo $USER && echo "$USER ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers && mkdir $INSTALLDIR && chown $USER:$USER $INSTALLDIR

# Installing our compiled server
USER $USER
COPY --chown=$USER:$USER --from=base $HOME/bin $INSTALLDIR/bin
COPY --chown=$USER:$USER --from=base $HOME/lib $INSTALLDIR/lib
COPY --chown=$USER:$USER docker_bootstrap.sh $INSTALLDIR/bin/bootstrap.sh
RUN echo "cd ~/" >> $HOME/.bashrc && chmod +x $INSTALLDIR/bin/bootstrap.sh && ln -s $INSTALLDIR/bin $HOME/bin && ln -s $INSTALLDIR/lib $HOME/lib && ln -s $INSTALLDIR/lib/logs $HOME/logs

# Expose master and slave ports
EXPOSE 28498
EXPOSE 28499

# Run our server, done!
ENTRYPOINT [ "/opt/gargoyle/bin/bootstrap.sh" ]
