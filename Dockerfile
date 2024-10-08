## -*- dockerfile-image-name: "logdy-core" -*-
FROM golang:1

WORKDIR /go/src/logdy-core/
COPY ./ /go/src/logdy-core/

RUN \
  --mount=type=cache,mode=0755,target=/root/.cache/go-build/ \
  --mount=type=cache,mode=0755,target=/go/pkg/mod/cache/ \
  go build -x -v \
  && go install -x -v \
  && go clean -v

ENTRYPOINT [ "/go/bin/logdy-core", "--ui-ip", "0.0.0.0" ]
CMD [ "stdin" ]
