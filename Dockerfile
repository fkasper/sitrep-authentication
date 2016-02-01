FROM busybox:ubuntu-14.04

MAINTAINER Florian Kasper <florian@xpandmmi.com>

# admin, http
EXPOSE 7717

WORKDIR /app

# copy binary into image
COPY build /app/build
COPY html /app/html
# Generate a default config
ADD sample.toml /etc/authentication.toml
ADD fargo.gcfg /etc/eureka.gcfg

ENTRYPOINT ["/app/build/authentication", "--config", "/etc/authentication.toml"]
