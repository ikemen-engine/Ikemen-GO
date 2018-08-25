# Ikemen GO Plus #

IKEMEN Go Plus is a fork of IKEMEN GO that aims to implement all remaining M.U.G.E.N features currently missing in the engine and add new ones on top of it in order to make the engine more flexible and better suited for full games.

IKEMEN Go is a remake of the IKEMEN (open source fighting games engine that supports M.U.G.E.N resources) in Google’s Programming Language “Go”.

## Building ##

### Windows ###

First, there are some programs to install before compiling.

**Git for Windows**: used to download this repository.

[https://gitforwindows.org/](https://gitforwindows.org/)

**Go/Golang**: used to compile golang code.

[https://golang.org/dl/](https://golang.org/dl/)

**TDM-GCC**: used to compile C++ code.

[http://tdm-gcc.tdragon.net/](http://tdm-gcc.tdragon.net/)


**OpenAL**: used to play sound.

[https://www.openal.org/](https://www.openal.org/)

After installing these programs, TDM-GCC needs some libraries to compile OpenAL code. So now, download OpenAL development libraries (**openal-soft-1.18.2-bin.zip**):

[http://kcat.strangesoft.net/openal.html](http://kcat.strangesoft.net/openal.html)

From that file, inside `include` folder, extract `AL` folder to TDM-GCC directory. By default, TDM-GCC is installed on `C:\TDM-GCC-64` (or 32) . The result should look like this:

![include directory result](https://vgy.me/oY3Zuk.png)

Also from that .zip file, inside `libs` folder, `libOpenAL32.dll.a` file should be extracted to TDM-GCC lib directory. By default it's in `C:\TDM-GCC-64\lib`. The result should look like this:

![lib directory result](https://vgy.me/c7FsG3.png)

After that, all the dependencies are installed and ready to do their work.

Now, download Ikemen GO Plus repository. It can be done downloading it as a zip from GitHub, or cloning the repository with Git. The latter is recommended to commit changes and then create a pull request.

Using Git:

`git clone https://github.com/K4thos/Ikemen-GO-Plus.git`

This will create a new folder with Ikemen code.

FINALLY, Ikemen can be compiled executing `build.bat` double clicking it or using cmd:

`./build.bat`

And now, Ikemen can be opened double clicking `Ikemen-GO-Plus.exe`

----------

### Linux ###

With a debian based system, it can be compiled executing the following commands on a terminal:

Install golang:

`sudo apt install golang-go`

Install git:

`sudo apt install git`

Install [GLFW](https://github.com/go-gl/glfw) dependencies:

`sudo apt install libgl1-mesa-dev xorg-dev`

Install OpenAL dependencies:

`sudo apt install libopenal1 libopenal-dev`

Download Ikemen GO Plus repository:

`git clone https://github.com/shinlucho/ikemen-plus.git`

Move to downloaded folder:

`cd ikemen-plus`

Execute get.sh to download Ikemen dependencies (it takes a while):

`./get.sh`

FINALLY compile:

`./build.sh`

And now, Ikemen can be opened double clicking Ikemen-GO-Plus, or with the terminal:

`./Ikemen_GO`
