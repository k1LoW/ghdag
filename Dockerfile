FROM golang:1.16-alpine

RUN apk add --no-cache bash

ENTRYPOINT ["/entrypoint.sh"]
CMD [ "-h" ]

COPY scripts/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

COPY ghdag_*.apk /tmp/
RUN apk add --allow-untrusted /tmp/ghdag_*.apk
