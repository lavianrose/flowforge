FROM python:3.12-alpine

RUN pip install --no-cache-dir requests httpx pyyaml jinja2

RUN adduser -D runner
USER runner

ENTRYPOINT ["/bin/sh", "-c", "echo \"$INPUTS_B64\" | base64 -d > /tmp/input.json && printf '%s' \"$CODE\" > /tmp/user_code.py && exec python3 /tmp/user_code.py"]
