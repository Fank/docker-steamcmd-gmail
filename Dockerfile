FROM ubuntu:14.04

MAINTAINER Florian Kinder <florian.kinder@fankserver.com>

# Install dependencies
RUN apt-get update &&\
	apt-get install --no-install-recommends --no-install-suggests -y curl lib32gcc1 &&\
	rm -rf /var/lib/apt/lists/*

# Download and extract SteamCMD
RUN mkdir -p /opt/steamcmd &&\
	cd /opt/steamcmd &&\
	curl -s http://media.steampowered.com/installer/steamcmd_linux.tar.gz | tar -vxz

WORKDIR /opt/steamcmd

# This container will be executable
ENTRYPOINT ["./steamcmd.sh"]
