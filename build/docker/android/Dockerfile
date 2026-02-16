ARG UBUNTU_VERSION=22.04
# Default uses Docker Hub; override BASE_IMAGE to use another registry (e.g. ECR Public)
ARG BASE_IMAGE=ubuntu:22.04
FROM ${BASE_IMAGE}

ARG DEBIAN_FRONTEND=noninteractive
ARG GO_VERSION=1.20.14
ARG ANDROID_NDK_VERSION=r27d
ARG ANDROID_NDK_ZIP=android-ndk-${ANDROID_NDK_VERSION}-linux.zip
ARG ANDROID_NDK_URL=https://dl.google.com/android/repository/${ANDROID_NDK_ZIP}

# Android SDK cmdline tools (can be overridden at build time if Google revs it)
ARG ANDROID_CMDLINE_TOOLS_ZIP=commandlinetools-linux-11076708_latest.zip
ARG ANDROID_CMDLINE_TOOLS_URL=https://dl.google.com/android/repository/${ANDROID_CMDLINE_TOOLS_ZIP}
ARG ANDROID_PLATFORM=android-34
ARG ANDROID_BUILD_TOOLS=34.0.0

RUN set -eux; \
  apt-get -o Acquire::Retries=5 update; \
  (apt-get -o Acquire::Retries=5 install -y --no-install-recommends \
    bash ca-certificates curl git unzip xz-utils zip file binutils \
    openjdk-17-jdk-headless build-essential make pkg-config cmake ninja-build \
    nasm yasm python3 \
  ) || ( \
    # Mirror got out of sync: refresh indexes and try once more
    apt-get update; \
    apt-get install -y --no-install-recommends \
      bash ca-certificates curl git unzip xz-utils zip file binutils \
      openjdk-17-jdk-headless build-essential make pkg-config cmake ninja-build \
      nasm yasm python3 \
  ); \
  rm -rf /var/lib/apt/lists/*

# Go (match CI's ~1.20)
RUN curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" -o /tmp/go.tgz \
  && rm -rf /usr/local/go \
  && tar -C /usr/local -xzf /tmp/go.tgz \
  && rm -f /tmp/go.tgz

ENV PATH=/usr/local/go/bin:${PATH}
ENV GOPATH=/go
ENV GOMODCACHE=/go/pkg/mod

# Android SDK (for Gradle APK build)
ENV ANDROID_SDK_ROOT=/opt/android-sdk
ENV ANDROID_HOME=/opt/android-sdk

RUN mkdir -p "${ANDROID_SDK_ROOT}/cmdline-tools" \
  && curl -fsSL "${ANDROID_CMDLINE_TOOLS_URL}" -o /tmp/cmdline-tools.zip \
  && unzip -q /tmp/cmdline-tools.zip -d "${ANDROID_SDK_ROOT}/cmdline-tools" \
  && rm -f /tmp/cmdline-tools.zip \
  && if [ -d "${ANDROID_SDK_ROOT}/cmdline-tools/cmdline-tools" ]; then \
       mv "${ANDROID_SDK_ROOT}/cmdline-tools/cmdline-tools" "${ANDROID_SDK_ROOT}/cmdline-tools/latest"; \
     fi

ENV PATH=${ANDROID_SDK_ROOT}/cmdline-tools/latest/bin:${ANDROID_SDK_ROOT}/platform-tools:${PATH}

RUN yes | sdkmanager --sdk_root="${ANDROID_SDK_ROOT}" --licenses >/dev/null \
  && sdkmanager --sdk_root="${ANDROID_SDK_ROOT}" \
      "platform-tools" \
      "platforms;${ANDROID_PLATFORM}" \
      "build-tools;${ANDROID_BUILD_TOOLS}"

# Android NDK r27d
RUN mkdir -p /opt \
  && curl -fsSL "${ANDROID_NDK_URL}" -o /tmp/ndk.zip \
  && unzip -q /tmp/ndk.zip -d /opt \
  && rm -f /tmp/ndk.zip \
  && test -d "/opt/android-ndk-${ANDROID_NDK_VERSION}"

ENV ANDROID_NDK_HOME=/opt/android-ndk-r27d

# Workspace (repo is mounted here via docker compose)
WORKDIR /work

# Avoid "dubious ownership" errors when the repo is volume-mounted
RUN git config --system --add safe.directory /work || true

CMD ["/bin/bash"]
