FROM alpine:3.13

LABEL description="free5GC UPF service" version="Stage 3"

ENV F5GC_MODULE=upf
ARG DEBUG_TOOLS

# Install bash + iptables + iproute2 + su-exec + debug tools if requested
RUN apk add --no-cache bash iptables iproute2 su-exec \
    && if [ "$DEBUG_TOOLS" = "true" ] ; then apk add --no-cache vim strace net-tools curl netcat-openbsd ; fi

# Create user/group and directories (do as root)
RUN addgroup -S free5gc && adduser -S free5gc -G free5gc \
    && mkdir -p /free5gc/config /free5gc/cert /free5gc/log

WORKDIR /free5gc

# Copy executable and scripts as root, ensure perms
COPY build/bin/${F5GC_MODULE} /free5gc/${F5GC_MODULE}
COPY config/upf-iptables.sh /free5gc/upf-iptables.sh
COPY config/upfcfg.yaml /free5gc/config/upfcfg.yaml

RUN chmod +x /free5gc/${F5GC_MODULE} /free5gc/upf-iptables.sh \
    && chown -R free5gc:free5gc /free5gc

# NOTE: keep USER as root so iptables can run in CMD; we will drop to free5gc using su-exec
# USER free5gc   <-- DO NOT set here

EXPOSE 8000

# Run iptables as root, then drop to free5gc to run UPF
CMD ["sh", "-c", "./upf-iptables.sh && su-exec free5gc ./upf -c ./config/upfcfg.yaml"]
