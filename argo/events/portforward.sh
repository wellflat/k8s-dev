#!/bin/sh

 kubectl port-forward svc/webhook-eventsource-svc 8080:8080 -n default
