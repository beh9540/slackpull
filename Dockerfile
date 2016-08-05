FROM golang:latest
MAINTAINER Blake Howell <sailpuff322@gmail.com>

# Set up workdir
WORKDIR /go/src/github.com/beh9540/slackpull

# Add and install slackpull
ADD . .
RUN go install -v github.com/beh9540/slackpull

# Expose the port
EXPOSE 8080

CMD ["slackpull"]