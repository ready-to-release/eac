FROM python:3.14-alpine

RUN apk add --no-cache git bash && \
    pip install --no-cache-dir git-filter-repo

WORKDIR /repo

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh && sed -i 's/\r$//' /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
