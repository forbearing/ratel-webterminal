#!/usr/bin/env bash


curl -i -N -H "Connection: Upgrade" -H "Upgrade: websocket" -H "Host: echo.websocket.org" -H "Origin: https://www.websocket.org" "$1"

#curl --include \
#     --no-buffer \
#     --header "Connection: Upgrade" \
#     --header "Upgrade: websocket" \
#     --header "Host: example.com:80" \
#     --header "Origin: http://example.com:80" \
#     --header "Sec-WebSocket-Key: SGVsbG8sIHdvcmxkIQ==" \
#     --header "Sec-WebSocket-Version: 13" \
#     http://localhost:9090/echo
