#!/bin/bash

if ! mongosh --eval 'rs.status()'; then
    if ! mongosh --eval "rs.initiate({_id: 'rs0', members: [ { _id: 0, host: '$WEBLENS_MONGO_HOST_NAME:27017' } ]})"; then
        echo "Failed to initiate replica set, exiting..."
        exit 1
    fi
fi

mongod --replSet rs0 --setParameter transactionLifetimeLimitSeconds=3600
