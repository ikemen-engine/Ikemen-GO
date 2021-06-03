#!/bin/bash
cd $(dirname $0)

case "$OSTYPE" in
	darwin*) #echo "It's a Mac!!" ;
		chmod +x Ikemen_GO_mac
		./IkemenGO_mac
		;;
	linux*)
		export MESA_GL_VERSION_OVERRIDE=2.1
		export MESA_GLES_VERSION_OVERRIDE=1.5
		chmod +x Ikemen_GO_linux
		./IkemenGO_linux
	;;
	*) echo "System not recognized"; exit 1 ;;
esac
