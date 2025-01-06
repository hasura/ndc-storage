#!/bin/sh

until (/usr/bin/mc config host add minio $STORAGE_ENDPOINT $ACCESS_KEY_ID $SECRET_ACCESS_KEY) 
do 
  echo '...waiting...' 
  sleep 1 
done 

/usr/bin/mc mb minio/$DEFAULT_BUCKET 
/usr/bin/mc policy set public minio/$DEFAULT_BUCKET