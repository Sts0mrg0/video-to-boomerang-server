FROM alpine:3.5

#RUN echo "http://dl-cdn.alpinelinux.org/alpine/edge/community/" >> /etc/apk/repositories
RUN apk add --update --no-cache ca-certificates ffmpeg graphicsmagick bash && \
    update-ca-certificates

WORKDIR /app

#ENTRYPOINT ["/app/main.sh"]
