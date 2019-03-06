# Part 1: Setting up the compiler (This is only required to do one time)

Step 1: download MSYS2 and follow the steps on the page.

https://www.msys2.org/

--------------------------------------------------------------------------------------------------------------------------------------------

Step 2: Write this onto the console to install mingw-w64 (The MSYS2 console)

for 32-bit:

`pacman -S mingw-w64-i686-gcc`

for 64-bit:

`pacman -S mingw-w64-x86_64-gcc`

Install the one based on your os version (Most computer today are 64-bit)

--------------------------------------------------------------------------------------------------------------------------------------------

Step 3: Install lib-png by writing this into the console (I don't know if this step is necessary but better safe than sorry)

for 32-bit:

`pacman -S mingw-w64-i686-libpng`

for 64-bit:

`pacman -S mingw-w64-x86_64-libpng`

--------------------------------------------------------------------------------------------------------------------------------------------

Step 4: Add the `<insert MSYS2 install folder here>/mingw64/bin` to your PATH

https://www.architectryan.com/2018/03/17/add-to-the-path-on-windows-10/

--------------------------------------------------------------------------------------------------------------------------------------------

Step 5: Download openal (openal-soft-1.19.1-bin.zip) Is the newest version at time of writing this]

http://kcat.strangesoft.net/openal.html

From that file, inside include folder, extract AL folder to the

`<insert MSYS2 install folder here>/mingw64//include` directory.

Also from that .zip file, inside libs folder, libOpenAL32.dll.a file should be extracted to

`<insert msys64 install folder here>/mingw64/lib directory`.

--------------------------------------------------------------------------------------------------------------------------------------------

Step 6: Install openAL

https://www.openal.org/downloads/

--------------------------------------------------------------------------------------------------------------------------------------------

Step 7: Install GO-lang (The newest version should work)

https://golang.org/

# Part 2: Buildidng the code

Download the repo.

Run the GET.cmd file

Run the BUILD.cmd file.
