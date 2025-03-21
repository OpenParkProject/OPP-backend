#!/bin/sh
set -e

echo "Getting API specs..."
rm -rf /tmp/OPP-common
git clone -b main --depth 1 https://github.com/OpenParkProject/OPP-common.git /tmp/OPP-common

echo "Generating API..."
cd src
mkdir -p api
cp -a /tmp/OPP-common/openapi.yaml api/openapi.yaml
go generate && echo "API generated" || echo "API generation failed"

echo "Building app..."
go build -buildvcs=false -o /go/bin/opp-backend .

echo "Starting app..."
/go/bin/opp-backend
