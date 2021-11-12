FROM golang:1.17.3-bullseye

ENV DEBIAN_FRONTEND=noninteractive
RUN wget -q -O - https://dl.google.com/linux/linux_signing_key.pub | apt-key add -
RUN sh -c 'echo "deb [arch=amd64] http://dl.google.com/linux/chrome/deb/ stable main" >> /etc/apt/sources.list.d/google.list'

RUN apt-get update
RUN apt-get upgrade -y 
RUN apt-get install -y google-chrome-stable

WORKDIR /app
COPY ./app .
RUN go build -o /bin/main
CMD ["main"]