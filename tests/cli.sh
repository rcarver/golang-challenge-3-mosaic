#!/bin/bash
# This script executes the mosaicly CLI to fetch images and generate a mosaic.
# It exits 0 on success. Any other exit is a failure.

set -e

# Write to stderr.
log() {
  echo "$@" 1>&2;
}

mosaicly=$GOPATH/bin/mosaicly
dir=`mktemp -d -t mosaicly`
log "Using tmp $dir"

log "Fetching..."
$mosaicly -run fetch -dir $dir -tag cat -num 5 
log "Testing tag dir"
test -d $dir/cat
log "Testing files in tag dir"
test 5 -eq $(ls $dir/cat/*.jpg | wc -l)

log "Generating..."
$mosaicly -run gen -dir $dir -tag cat -in fixtures/balloon.jpg -units 10 -out $dir/test.jpg
log "Testing output"
test -f $dir/test.jpg

log "Opening image..."
open $dir/test.jpg

