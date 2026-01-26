# Static Site Container
# Lightweight nginx image for serving pre-built static files
# ~25MB image size

FROM nginx:alpine

# Remove default nginx config
RUN rm /etc/nginx/conf.d/default.conf

# Copy custom nginx config
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Create directory for site content
RUN mkdir -p /usr/share/nginx/html

EXPOSE 8000

CMD ["nginx", "-g", "daemon off;"]
