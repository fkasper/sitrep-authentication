FROM busybox:ubuntu-14.04

MAINTAINER Florian Kasper <florian@xpandmmi.com>

# admin, http
EXPOSE 7717, 7727

WORKDIR /app

# copy binary into image
COPY bio /app/

# Add influxd to the PATH
ENV PATH=/app:$PATH

# Generate a default config
RUN bio config > /etc/bio.toml

# Use /data for all disk storage
RUN sed -i 's/dir = "\/.*bio/dir = "\/data/' /etc/bio.toml

VOLUME ["/data"]

ENTRYPOINT ["bio", "--config", "/etc/bio.toml"]
