# build stage
FROM golang:alpine AS build-env
#RUN apk --no-cache add build-base git gcc
ADD . /src
RUN cd /src/cmd/server && go build -o main

# final stage
FROM alpine
WORKDIR /app
COPY --from=build-env /src/cmd/server/main /app/main
COPY --from=build-env /src/configs/docker_config.toml /app/configs/docker_config.toml
ENTRYPOINT ./main -c ./configs/docker_config.toml
