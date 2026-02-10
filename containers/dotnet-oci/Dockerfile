# .NET SDK Container
# General-purpose container for .NET builds, tests, and toolchain operations.
# Used by the tool system as an alternative to native 'dotnet' binary.
#
# Features:
# - .NET 8 LTS SDK (build, test, publish, format)
# - NuGet restore with shared package cache (mount /root/.nuget/packages)
# - Cross-platform publish support via --runtime flag
# - xUnit, NUnit, MSTest runner support
# - Reqnroll BDD test execution
#
# Usage via tool system:
#   ArgsOverrides: ["build", "--configuration", "Release"]
#   ArgsOverrides: ["test", "--logger", "trx"]
#   ArgsOverrides: ["publish", "--runtime", "linux-x64", "--self-contained"]

ARG DOTNET_VERSION=8.0
FROM mcr.microsoft.com/dotnet/sdk:${DOTNET_VERSION}-alpine

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache \
    git \
    ca-certificates \
    icu-libs \
    && rm -rf /var/cache/apk/*

# Disable .NET telemetry in container
ENV DOTNET_CLI_TELEMETRY_OPTOUT=1
ENV DOTNET_NOLOGO=1

# Enable globalization support (needed for some test frameworks)
ENV DOTNET_SYSTEM_GLOBALIZATION_INVARIANT=false

# Pre-create NuGet cache directory (will be mounted from host)
RUN mkdir -p /root/.nuget/packages

# Entrypoint is 'dotnet' - all subcommands come from ArgsOverrides
ENTRYPOINT ["dotnet"]

# Default command if no ArgsOverrides provided
CMD ["--version"]
