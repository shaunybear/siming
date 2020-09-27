FROM golang:1.15.2-buster

ARG USER_ID=1000
ARG GROUP_ID=1000

RUN addgroup --gid $GROUP_ID simmer
RUN adduser --disabled-password --gecos '' --uid $USER_ID --gid $GROUP_ID simmer
USER simmer