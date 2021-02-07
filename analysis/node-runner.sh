#!/bin/bash

for pkg in $(cat node.txt); do
    kubectl run $pkg --image=node --restart='Never' --labels=install=1 --requests="cpu=250m" -- sh -c "npm init -f && npm install $pkg --save"
done
