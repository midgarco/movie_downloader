FROM golang:buster
RUN apt-get update && apt-get install -y git curl openssh-client
RUN git config --global url."git@github.com:".insteadOf "https://github.com/"
RUN git config --global url."git://".insteadOf "https://"
RUN git config -l
RUN mkdir -p /root/.ssh

WORKDIR /src
COPY . /src

ARG PRIVATE_ID_KEY
RUN echo "${PRIVATE_ID_KEY}" | head -c 42
RUN echo "${PRIVATE_ID_KEY}" | tail -c 42
RUN date && echo "${PRIVATE_ID_KEY}" > /root/.ssh/id_rsa && chmod 400 /root/.ssh/id_rsa && touch /root/.ssh/known_hosts && ssh-keyscan github.com >> /root/.ssh/known_hosts
RUN go clean -modcache
RUN rm -f pmd-server
RUN go build -o pmd-server -v ./cmd/server

RUN rm /root/.ssh/id_rsa 

RUN mkdir -p /download
RUN mkdir -p /media

EXPOSE 4050
EXPOSE 4051

CMD ["./pmd-server", "--d", "/download", "--media", "/media"]
