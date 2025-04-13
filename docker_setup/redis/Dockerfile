# Use the official lightweight Redis image.
FROM redis:6.2-alpine

# Copy the custom configuration file into the image.
COPY redis.conf /usr/local/etc/redis/redis.conf

# Start Redis using the custom configuration file.
CMD ["redis-server", "/usr/local/etc/redis/redis.conf"]