# Use the Ubuntu 24.04 base image
FROM ubuntu:24.04

# Set environment variables for non-interactive installs
ENV DEBIAN_FRONTEND=noninteractive

# Update and install required packages
RUN apt-get update && apt-get install -y \
    curl \
    unzip \
    bash \
    ca-certificates && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Define the environment directory for tfvenv
ENV TFVENV_DIR=/tfvenvroot

# Create a volume for tfvenvroot so it can be mounted from the host
VOLUME /tfvenvroot

# Copy the tfvenv configuration sample if it exists
# Note: Modify this step if you want to include a sample .tfvenvrc file in the image
COPY .tfvenvrc.sample /tfvenvroot/.tfvenvrc

# Add an entrypoint to source the activation script and bring up a bash session
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

# Set the default command to run the entrypoint
ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]

# Default to running bash
CMD ["/bin/bash"]
