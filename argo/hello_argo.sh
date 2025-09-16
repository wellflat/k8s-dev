#!/bin/sh

argo submit --watch -n default --serviceaccount=workflow-sa hello-world.yaml
