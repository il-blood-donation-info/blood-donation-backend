# Use a lightweight base image
FROM alpine:latest

RUN apk --no-cache add jq

# Set the working directory
WORKDIR /app

# Copy the test.sh script into the container
COPY test/integration/test.sh /app/test.sh

# Make the script executable
RUN chmod +x ./test.sh

# Define the command to run the test
CMD ["./test.sh"]