FROM scratch

COPY build/malzone /malzone

USER 65532:65532
ENTRYPOINT ["/malzone"]
