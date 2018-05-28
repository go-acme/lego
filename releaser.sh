#!/usr/bin/env bash
set -e

if [ -z "$VERSION" ]; then
    VERSION=$(git describe --long --dirty=+ --abbrev=12 --tags)
fi

BUILD_FOLDER="builds"
DIST_FOLDER="dist"

rm -rf ${DIST_FOLDER}
mkdir ${DIST_FOLDER}

GO_BUILD_CMD="go build -ldflags"
GO_BUILD_OPT="-s -w -X main.gittag=${VERSION}"

buildPackage() {
    for OS in ${OS_PLATFORM_ARG[@]}; do
      BIN_EXT=''
      if [ "$OS" == "windows" ]; then
        BIN_EXT='.exe'
      fi
      for ARCH in ${OS_ARCH_ARG[@]}; do

        rm -rf ${BUILD_FOLDER}
        mkdir ${BUILD_FOLDER}

        cp LICENSE ${BUILD_FOLDER}/LICENSE.txt
        cp README.md ${BUILD_FOLDER}/README.md
        cp CHANGELOG.md ${BUILD_FOLDER}/CHANGELOG.md

        echo "Building binary for ${OS}/${ARCH}..."
        GOARCH=${ARCH} GOOS=${OS} CGO_ENABLED=0 ${GO_BUILD_CMD} "${GO_BUILD_OPT}" -o "${BUILD_FOLDER}/lego_${OS}_${ARCH}${BIN_EXT}"
        cd builds
        if [ "$OS" == "windows" ]; then
            zip -r ../dist/lego_${OS}_${ARCH}.zip .
        else
            tar -czvf ../dist/lego_${OS}_${ARCH}.tar.gz .
        fi
        cd ..
      done
    done
}

# Build linux binaries
OS_PLATFORM_ARG=(linux)
OS_ARCH_ARG=(amd64 386 arm64 arm)
buildPackage

# Build freebsd binaries
OS_PLATFORM_ARG=(freebsd)
OS_ARCH_ARG=(amd64 386 arm)
buildPackage

# Build openbsd binaries
OS_PLATFORM_ARG=(openbsd)
OS_ARCH_ARG=(amd64 386)
buildPackage

# Build darwin binaries
OS_PLATFORM_ARG=(darwin)
OS_ARCH_ARG=(amd64)
buildPackage

# Build windows binaries
OS_PLATFORM_ARG=(windows)
OS_ARCH_ARG=(amd64)
buildPackage

rm -rf ${BUILD_FOLDER}
