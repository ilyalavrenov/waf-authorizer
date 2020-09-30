FROM golang as gobuild
WORKDIR /go/src/app
COPY . .
RUN GOOS=linux go build -ldflags "-linkmode external -extldflags -static" -o /web main.go

FROM alpine:latest as alpine
RUN apk --no-cache add tzdata zip ca-certificates gzip curl
WORKDIR /usr/share/zoneinfo
RUN zip -r -0 /zoneinfo.zip .
ENV GCREDSTASH_VERSION="0.3.5" \
    GCREDSTASH_SHA256="462c78808c2dd2a87db61ee87228fbf53f9b638c14b1dc96fa10947ad444f579"
WORKDIR /tmp
RUN curl -sLo gcredstash.gz https://github.com/winebarrel/gcredstash/releases/download/v${GCREDSTASH_VERSION}/gcredstash-v${GCREDSTASH_VERSION}-linux-amd64.gz && \
    echo "${GCREDSTASH_SHA256}  gcredstash.gz" | sha256sum -c - && gunzip gcredstash.gz && \
    mv gcredstash /usr/local/bin/gcredstash && chmod +x /usr/local/bin/gcredstash

FROM nginx:alpine as nginx
COPY ./nginx/nginx.conf /etc/nginx/nginx.conf
COPY ./nginx/conf.d/default.conf /etc/nginx/conf.d/default.conf

FROM busybox
COPY docker-entrypoint.sh /docker-entrypoint.sh
COPY --from=gobuild /web /web
ENV ZONEINFO /zoneinfo.zip
COPY --from=alpine /zoneinfo.zip /
COPY --from=alpine /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=alpine /usr/local/bin/gcredstash /usr/local/bin/gcredstash
COPY --from=nginx /etc/nginx /etc/nginx
VOLUME ["/etc/nginx"]
ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["/web"]
