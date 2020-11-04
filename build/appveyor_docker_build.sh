cd ..

if [ ! -d ./bin ]; then
	mkdir bin
fi

echo "------------------------------------------------------------"
echo "Building Proccess starting."
docker run --rm -v $(pwd):/code -i windblade/ikemen-dev:latest bash -c 'cd /code/build && bash -x appveyor_build_step2.sh' 

# We copy the Windres files so we can have a icon files
cp 'windres/Ikemen_Cylia_x64.syso' 'src/Ikemen_Cylia_x64.syso'

echo "------------------------------------------------------------"
echo "All builds finished."
echo "------------------------------------------------------------"