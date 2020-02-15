FROM arm32v7/alpine
COPY ./customerd /bin/customerd
RUN ["chmod", "+x", "/bin/customerd"]
ENTRYPOINT ["/bin/customerd"]