FROM arm32v7/alpine

RUN apk add --no-cache curl

COPY ./customerd /bin/customerd
RUN ["chmod", "+x", "/bin/customerd"]
ENTRYPOINT ["/bin/customerd"]