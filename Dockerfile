FROM golang:1.15.2-buster

# Avoid warnings by switching to noninteractive
ENV DEBIAN_FRONTEND=noninteractive

ARG USER_ID=1000
ARG GROUP_ID=1000
ARG USERNAME=simmer

# Configure apt and install packages
RUN apt-get update \
    && apt-get -y install --no-install-recommends apt-utils dialog 2>&1 \
    # Create a no-root user to use if preferred
    && addgroup --gid $GROUP_ID simmer \
    && adduser --disabled-password --gecos '' --uid $USER_ID --gid $GROUP_ID $USERNAME \
    # Add sudo support for the user
    && apt-get -y install  sudo \ 
    && echo  $USERNAME ALL=\(root\) NOPASSWD:ALL > /etc/sudoers.d/$USERNAME \
    && chmod 0440 /etc/sudoers.d/$USERNAME \
    # Cleanup
    && apt-get autoremove -y \
    && apt-get clean -y \
    && rm -rf /var/lib/apt/lists/*


USER simmer