#/usr/bin/env bash
set -ce

dotnet restore "build-src"  \
  --packages "build-cache"

dotnet publish "build-src"    \
  --packages "build-cache"  \
  --configuration "{{build_configuration}}"
  --output "$(pwd)/{{product_name}}"

tar czf "build-bin/{{product_name}}.tar.gz" {{product_name}}/*
