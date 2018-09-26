#/usr/bin/env bash
set -ce

dotnet restore "build-src"  \
  --packages "build-cache"

dotnet build "build-src"    \
  --packages "build-cache"  \
  --configuration "Debug"

dotnet test "build-src"     \
  --no-build                \
  --no-restore
