### Copyright (c) 2020 Richard Youngkin. All rights reserved.
### Use of this source code is governed by a MIT-style
### license that can be found in the LICENSE file.

FROM arm32v7/alpine

RUN apk add --no-cache curl

COPY ./accountd /bin/accountd
RUN ["chmod", "+x", "/bin/accountd"]
ENTRYPOINT ["/bin/accountd"]