FROM gcr.io/distroless/static-debian11:nonroot
ENTRYPOINT ["/baton-zuper"]
COPY baton-zuper /