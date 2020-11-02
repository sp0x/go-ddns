#!/bin/bash
# Use this script to make calls to a cloud function.
# This way it's really easy update your records

# Domain can be a FQDN or just a partial name.
curl "${ENDPOINT}?secret=${API_KEY}&domain=${SUBDOMAIN}"

#If you also need your local address to be added
#install the npm module internal-ip-cli and execute
curl "${ENDPOINT}?secret=${API_KEY}&domain=local.${SUBDOMAIN}&addr=$(internal-ip --ipv4)"
