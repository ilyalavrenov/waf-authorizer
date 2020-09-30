#!/bin/sh

export API_KEY=$(gcredstash get waf-authorizer-api-key-${ENVIRONMENT})

if test "$API_KEY" == ""; then echo "unable to fetch api key"; exit 1; fi

exec "$@"