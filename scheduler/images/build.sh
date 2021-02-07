#!/bin/bash

docker build . -f Dockerfile.python -t gcr.io/ossf-malware-analysis/python
docker build . -f Dockerfile.node -t gcr.io/ossf-malware-analysis/node
docker build . -f Dockerfile.ruby -t gcr.io/ossf-malware-analysis/ruby

docker push gcr.io/ossf-malware-analysis/python
docker push gcr.io/ossf-malware-analysis/node
docker push gcr.io/ossf-malware-analysis/ruby
