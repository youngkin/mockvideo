FROM arm32v7/alpine
COPY cmd/customerd/customerd /bin/customerd
RUN ["chmod", "+x", "/bin/customerd"]
ENTRYPOINT ["/bin/customerd"]