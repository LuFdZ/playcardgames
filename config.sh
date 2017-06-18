#!/usr/bin/env bash

curl -XPUT http://localhost:8500/v1/kv/playcards/config -d '{
    "log_level": "debug",
    "db_url": "root:@tcp(127.0.0.1:3306)/playcards?parseTime=true&loc=Asia%2FChongqing",
    "redis_host": "127.0.0.1:6379",
    "api_host": "0.0.0.0:8080",
    "web_host": "0.0.0.0:8999"
}'
