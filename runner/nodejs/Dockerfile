FROM node:20-alpine

RUN npm install -g axios lodash && mkdir -p /runner

WORKDIR /runner

COPY run.sh /runner/run.sh
RUN chmod +x /runner/run.sh

RUN adduser -D runner
USER runner

ENTRYPOINT ["/runner/run.sh"]
