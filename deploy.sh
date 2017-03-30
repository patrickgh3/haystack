#!/bin/bash
cp $GOPATH/bin/haystack /home/deploy/haystack/haystack
cp -r $GOPATH/src/github.com/patrickgh3/haystack/templates /home/deploy/haystack
sudo cp -r $GOPATH/src/github.com/patrickgh3/haystack/assets/* /var/html/cwpat.me/haystack
