cd ..

if [ ! -d ./bin ]; then
	mkdir bin
fi

echo "------------------------------------------------------------"
docker run --rm -e OS=linux -v $(pwd):/code -i windblade/ikemen-dev:latest bash -c 'cd /code/build  && bash -x get.sh'

echo "------------------------------------------------------------"
echo "Building linux binary..."
docker run --rm -e OS=linux -v $(pwd):/code -i windblade/ikemen-dev:latest bash -c 'cd /code/build && bash -x build_crossplatform.sh' 

echo "------------------------------------------------------------"
echo "Building mac binary..."
docker run --rm -e OS=mac -v $(pwd):/code -i windblade/ikemen-dev:latest bash -c 'cd /code/build && bash -x build_crossplatform.sh' 

echo "------------------------------------------------------------"
echo "Building windows x86 binary..."
docker run --rm -e OS=windows32 -v $(pwd):/code -i windblade/ikemen-dev:latest bash -c 'cd /code/build && bash -x build_crossplatform.sh' 

# We copy the Windres files so we can have a icon files
cp 'windres/Ikemen_Cylia_x64.syso' 'src/Ikemen_Cylia_x64.syso'

echo "------------------------------------------------------------"
echo "Building windows x64 binary..."
docker run --rm -e OS=windows -v $(pwd):/code -i windblade/ikemen-dev:latest bash -c 'cd /code/build && bash -x build_crossplatform.sh' 
