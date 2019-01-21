# Ikemen GO
IKEMEN Go is a remake of the IKEMEN (open source fighting games engine that supports M.U.G.E.N resources) in Google’s Programming Language “Go”.

### Links ###

Suehiro repo (Original creator of the engine)
https://osdn.net/users/supersuehiro/pf/ikemen_go/

Ikemen GO Plus, K4thos fork of Ikemen (Some features were borrowed from his repo)
https://github.com/K4thos/Ikemen-GO-Plus

# Building #

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

`git clone https://github.com/K4thos/Ikemen-GO-Plus.git`

Move to downloaded folder:

`cd Ikemen-GO-Plus`

Execute get.sh to download Ikemen dependencies (it takes a while):

`./get.sh`

FINALLY compile:

`./build.sh`

And now, Ikemen can be opened double clicking Ikemen-GO-Plus, or with the terminal:

`./Ikemen_GO`
