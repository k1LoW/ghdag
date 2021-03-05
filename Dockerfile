FROM alpine:3.13

RUN apk add --no-cache bash curl git

ENTRYPOINT ["/entrypoint.sh"]
CMD [ "-h" ]

COPY scripts/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

COPY ghdag_*.apk /tmp/
RUN apk add --allow-untrusted /tmp/ghdag_*.apk
