FROM valhalla/valhalla:run-latest

RUN apt update
RUN apt --assume-yes install wget

COPY build.sh build.sh

RUN chmod +x build.sh

ENTRYPOINT ["/build.sh"]