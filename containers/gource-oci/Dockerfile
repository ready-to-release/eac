# Gource Visualization Container
# Headless Gource with HLS live streaming
ARG ALPINE_VERSION=3.20
FROM alpine:${ALPINE_VERSION}

# Install dependencies:
# - gource: visualization tool
# - xvfb: virtual framebuffer for headless display
# - mesa-gl, mesa-dri-gallium: software OpenGL rendering
# - ffmpeg: video encoding to HLS
# - nginx: serve HLS stream
# - git: required by gource
RUN apk add --no-cache \
    gource \
    xvfb \
    mesa-gl \
    mesa-dri-gallium \
    ffmpeg \
    nginx \
    git \
    bash \
    font-dejavu

# Create directories
RUN mkdir -p /var/run/nginx /visualization/hls /visualization/repo /visualization/output /visualization/html

# Copy nginx config for HLS streaming
COPY nginx.conf /etc/nginx/nginx.conf

# Copy HTML player
COPY index.html /visualization/html/index.html

# Copy entrypoint script and fix Windows line endings
COPY entrypoint.sh /entrypoint.sh
RUN sed -i 's/\r$//' /entrypoint.sh && chmod +x /entrypoint.sh

EXPOSE 80

WORKDIR /visualization/repo

ENV GOURCE_TITLE="Repository History"
ENV GOURCE_RESOLUTION="960x540"
ENV GOURCE_SECONDS_PER_DAY="0.5"

ENTRYPOINT ["/entrypoint.sh"]
