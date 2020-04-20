FROM alpine:latest
ADD terrafmt /bin/

ENTRYPOINT ["/bin/sh"]
