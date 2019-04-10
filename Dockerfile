FROM golang:1.11-alpine as builder

MAINTAINER Chakib <contact@bitcointechweekly.com>

# Copy source
COPY . /go/src/hugobot

# install dependencies and build
RUN apk add --no-cache --upgrade \
    ca-certificates \
    git \
    openssh \
    make \
    alpine-sdk

RUN cd /go/src/hugobot \
&&  make install

################################
#### FINAL IMAGE
###############################


FROM alpine as final

ENV WEBSITE_PATH=/website
ENV HUGOBOT_DB_PATH=/db

RUN apk add --no-cache --upgrade \
    ca-certificates \
    bash \
    sqlite \
    jq

COPY --from=builder /go/bin/hugobot /bin/


RUN mkdir -p ${HUGOBOT_DB_PATH}
RUN mkdir -p ${WEBSITE_PATH}


VOLUME ${HUGOBOT_DB_PATH}


# Expose API ports
EXPOSE 8734

# copy entrypoint
COPY "docker-entrypoint.sh" /entry

ENTRYPOINT ["/entry"]
CMD ["hugobot", "server"]
