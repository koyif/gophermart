FROM ubuntu:24.04

RUN apt-get update &&  \
    apt-get install -y libc6-amd64-cross  \
    && ln -s /usr/x86_64-linux-gnu/lib64/ /lib64 \
    && apt-get clean && rm -rf /var/lib/apt/lists/*

ENV LD_LIBRARY_PATH="$LD_LIBRARY_PATH:/lib64:/usr/x86_64-linux-gnu/lib"

WORKDIR /
COPY --chmod=0755 accrual_linux_amd64 /accrual_linux_amd64
USER 1000

EXPOSE 8080

ENTRYPOINT ["/accrual_linux_amd64"]