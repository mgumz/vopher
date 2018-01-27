FROM alpine:3.7 as build

ARG  VOPHER
ARG  BUILD_DIR

RUN  apk --update add go make musl-dev git
COPY . $BUILD_DIR
RUN  make -C $BUILD_DIR $VOPHER

###

FROM alpine:3.7
ARG  VOPHER
ARG  BUILD_DIR
RUN  apk --update add ca-certificates
COPY --from=build $BUILD_DIR/$VOPHER /usr/bin/vopher
