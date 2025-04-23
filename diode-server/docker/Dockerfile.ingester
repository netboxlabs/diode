FROM gcr.io/distroless/base:latest

USER nonroot

ARG SVC

LABEL maintainer="techops@netboxlabs.com"
LABEL org.opencontainers.image.title="Diode $SVC"
LABEL org.opencontainers.image.url="https://github.com/netboxlabs/diode"
LABEL org.opencontainers.image.documentation="https://github.com/netboxlabs/diode"
LABEL org.opencontainers.image.description="NetBox Labs diode $SVC image"
LABEL org.opencontainers.image.vendor="NetBox Labs, Inc"
LABEL org.opencontainers.image.authors="techops@netboxlabs.com"
LABEL org.opencontainers.image.licenses="PolyForm Shield License 1.0.0"

COPY build/$SVC /usr/local/bin/diode-server

CMD ["/usr/local/bin/diode-server"]
