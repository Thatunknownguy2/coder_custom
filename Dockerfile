FROM alpine:latest
RUN adduser -D -h /home/coder -s /bin/bash coder
RUN apk update; apk add vim bash
USER coder
COPY build/coder_0.17.4-devel+????????_linux_arm64 /home/coder/coder
WORKDIR /home/coder/
ENTRYPOINT ["./coder", "server"]

# An example
# docker run --rm -it -p 4000:4000 -e CODER_ACCESS_URL="http://localhost:4000" coder_custom
