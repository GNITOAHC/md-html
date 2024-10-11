#!/bin/bash

VERSION=$(git describe --tags --abbrev=0)
BINARY="md-html"
BUILD_DIR="build"

mkdir -p $BUILD_DIR

PLATFORMS=("linux/amd64" "darwin/amd64" "darwin/arm64" "windows/amd64")

for platform in "${PLATFORMS[@]}"; do
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}

    output_name=$BINARY
    if [ $GOOS = "windows" ]; then
        output_name+='.exe'
    fi

    GOOS=$GOOS GOARCH=$GOARCH go build -o $BUILD_DIR/$output_name -ldflags "-s -w -X main.Version=$VERSION"
    tar -czf $BUILD_DIR/$BINARY-$VERSION-$GOOS-$GOARCH.tar.gz -C $BUILD_DIR $output_name

    echo "Build $BINARY-$VERSION-$GOOS-$GOARCH.tar.gz done"
done

rm $BUILD_DIR/$BINARY
rm $BUILD_DIR/$BINARY.exe
