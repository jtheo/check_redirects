#!/bin/bash

function log() {
  echo "$(date): ========================> ${*}"
}

function err() {
  printf "${*}" | xargs -IX bash -c 'echo -e "$(date): \033[0;31m X\033[0m"'
  exit 1
}

log "Running go vet"
if ! go vet ./...; then
  echo "go vet failed"
  exit 1
fi

shopt -s nullglob
set -- *_test.go
if [ "$#" -gt 0 ]; then
  log "Running go test"
  if ! go test -v; then
    echo "go test failed"
    exit 1
  fi
fi
shopt -u nullglob

if type -p govulncheck >/dev/null; then
  log "Running govulncheck"
  govulncheck ./...
fi

if type -p golangci-lint >/dev/null; then
  log "Running golangci-lint"
  if ! golangci-lint run -E revive -E errcheck -E nilerr -E gosec -E staticcheck -E prealloc -E nilerr -E gochecksumtype -E exhaustruct; then
    err "Check above..."
  fi
fi

D=$(basename "${PWD}")
dst=dist
verFile=version.txt
name=${1:-$D}
mkdir -p "${dst}"

if [[ -e ${verFile} ]]; then
  version=$(tr -d '\n' <${verFile})
  version=$((version + 1))
else
  version=1
fi

oses=(linux darwin)
archs=(amd64 arm64)

log "building version ${version}..."
echo "Building "

for GOOS in "${oses[@]}"; do
  printf "  - %s " "${GOOS}"
  for GOARCH in "${archs[@]}"; do
    printf "%s " "${GOARCH}"
    # shellcheck disable=SC2097,SC2098
    GOOS=${GOOS} GOARCH=${GOARCH} CGO_ENABLED=0 \
      go build -ldflags "-s -w -X 'main.Version=${version}'" \
      -o "${dst}/${name}-${GOOS}-${GOARCH}" .
  done
  echo
done

printf "%s" "${version}" >"${verFile}"
echo
