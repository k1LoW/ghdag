FROM debian:buster-slim

RUN apt-get update && apt-get install -y \
  curl \
  git \
 && apt-get clean \
 && rm -rf /var/lib/apt/lists/*

ENTRYPOINT ["/entrypoint.sh"]
CMD [ "-h" ]

COPY scripts/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

COPY ghdag_*.deb /tmp/
RUN dpkg -i /tmp/ghdag_*.deb
