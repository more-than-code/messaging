user=${1}

docker buildx build --platform linux/arm64 -t ${user}/messaging:latest --push .