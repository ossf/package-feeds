#!/bin/bash

for pkg in $(cat python.txt); do
    kubectl run $pkg --image=python:3 --restart='Never' --labels=install=1 --requests="cpu=250m" -- sh -c "pip3 install $pkg"
done
