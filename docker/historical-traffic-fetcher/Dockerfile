FROM curlimages/curl:7.85.0

COPY fetch.sh fetch.sh

RUN chmod +x fetch.sh

ENTRYPOINT ["/fetch.sh"]