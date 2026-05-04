FROM python:3.12-alpine

RUN pip install --no-cache-dir requests httpx pyyaml jinja2

WORKDIR /runner

COPY run.sh /runner/run.sh
RUN chmod +x /runner/run.sh

RUN adduser -D runner
USER runner

ENTRYPOINT ["/runner/run.sh"]
