#!/bin/bash
curl "${ENDPOINT}?secret=${API_KEY}&domain=${SUBDOMAIN}"

#If you also need your local address to be added
#install the npm module internal-ip-cli and execute
#curl "${ENDPOINT}?secret=${API_KEY}&domain=local.${SUBDOMAIN}&addr=$(internal-ip --ipv4)"
