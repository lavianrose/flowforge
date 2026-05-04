FROM node:20-alpine

RUN npm install -g axios lodash

RUN adduser -D runner
USER runner

ENTRYPOINT ["/bin/sh", "-c", "echo \"$INPUTS_B64\" | base64 -d > /tmp/input.json && printf '%s' \"$CODE\" > /tmp/user_code.js && exec node /tmp/user_code.js"]
