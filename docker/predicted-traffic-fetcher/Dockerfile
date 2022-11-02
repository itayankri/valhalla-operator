FROM valhalla/valhalla:run-latest

COPY fetch.sh fetch.sh

RUN chmod +x fetch.sh

ENTRYPOINT ["/fetch.sh"]