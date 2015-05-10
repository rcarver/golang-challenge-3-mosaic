#!/bin/bash

set -e

endpoint="localhost:8080"
tag="balloon"
img="balloon.jpg"

# Create a new mosaic
res="$(curl -fs -X POST -F img=@fixtures/${img} "$endpoint/mosaics?tag=${tag}")"
echo $res
id=$(echo $res | jq -M -r .id)
url=$(echo $res | jq -M -r .url)
img_url=$(echo $res | jq -M -r .img)
status=$(echo $res | jq -M -r .status)

# Show the URL
echo "new one id:$id status:$status at ${url}"

# Wait for the image to be done.
while [[ "$status" != "created" ]]; do
  status=$(curl -fs "${endpoint}${url}" | jq -M -r .status)
  echo "waiting for mosaic..."
  sleep 1
done;

# Show the mosaics list.
res=$(curl -fs "$endpoint/mosaics")
echo $res | jsonpretty

# Open the image.
echo "opening image..."
curl -fs ${endpoint}${img_url} > /tmp/${img}
open /tmp/${img}

