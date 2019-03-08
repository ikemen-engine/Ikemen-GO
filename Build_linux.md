# Building on Linux

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

`git clone https://github.com/Windblade-GR01/Ikemen_GO.git`

Move to downloaded folder:

`cd Ikemen-GO-Plus`

Execute get.sh to download Ikemen dependencies (it takes a while):

`./get.sh`

FINALLY compile:

`./build.sh`

And now, Ikemen can be opened double clicking Ikemen-GO-Plus, or with the terminal:

`./Ikemen_GO`
