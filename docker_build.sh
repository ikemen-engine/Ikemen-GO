echo "Building linux binary..."
docker run --rm -e OS=linux -v $(pwd):/code -it danielporto/ikemen-dev:latest bash -c 'cd /code && bash -x get.sh' 
docker run --rm -e OS=linux -v $(pwd):/code -it danielporto/ikemen-dev:latest bash -c 'cd /code && bash -x build.sh' 

echo "Building windows binary..."
docker run --rm -e OS=windows -v $(pwd):/code -it danielporto/ikemen-dev:latest bash -c 'cd /code && bash -x get.sh' 
docker run --rm -e OS=windows -v $(pwd):/code -it danielporto/ikemen-dev:latest bash -c 'cd /code && bash -x build.sh' 

echo "Building mac binary..."
docker run --rm -e OS=mac -v $(pwd):/code -it danielporto/ikemen-dev:latest bash -c 'cd /code && bash -x get.sh' 
docker run --rm -e OS=mac -v $(pwd):/code -it danielporto/ikemen-dev:latest bash -c 'cd /code && bash -x build.sh' 