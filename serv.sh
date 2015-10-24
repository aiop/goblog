#!/bin/bash
case "$1" in
'start')
nohup ./gowebserver
echo "start...."
;;
'stop')
pkill ./gowebserver
echo "stop...."
;;
'restart')
pkill gowebserver
nohup ./gowebserver
echo "start...."
;;
*)
;;
esac
exit 0
