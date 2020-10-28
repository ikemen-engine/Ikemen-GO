#!/bin/bash
cd $(dirname $0)

case "$OSTYPE" in
    darwin*) #echo "It's a Mac!!" ;
        chmod +x Ikemen_GO_mac
        ./Ikemen_GO_mac
        ;;
    linux*) 
        chmod +x Ikemen_GO_linux
        ./Ikemen_GO_linux
     ;;
    *) echo "System not recognized"; exit 1 ;;
esac