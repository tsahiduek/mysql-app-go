
# ARG binary_name
ARG alpine_version=3.12

FROM alpine:${alpine_version}
COPY ./form/* /form/
COPY ./dist/linux/octicketing /octicketing
ENTRYPOINT [ "/octicketing" ]
