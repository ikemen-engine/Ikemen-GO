# Ikemen GO
IKEMEN Go is a remake of the IKEMEN (open source fighting games engine that supports M.U.G.E.N resources) in Google’s Programming Language “Go”.

## Installing
Ready to use builds for Windows, MacOS and Linux can be found on the releases tab of the repo. 

## Running
On windows, execute `Ikemen_GO.exe` (`Ikemen_GO_x86.exe` on 32-bit OS)  
On MacOS or Linux, double-click on `Ikemen_GO.command`

## Developing
These instructions are for those interested in developing the Ikemen_GO engine. Instructions on contributing with custom stages, fonts, characters and other resources can be found in the community forum.

### Building on Windows
Check the instructions [here](https://github.com/Windblade-GR01/Ikemen_GO/wiki/Building-on-Windows)

### Building on Mac
Check the insturctions [here](https://github.com/Windblade-GR01/Ikemen_GO/wiki/Building-on-MacOS)

### Building on Linux
Check the instructions [here](https://github.com/Windblade-GR01/Ikemen_GO/wiki/Building-on-Linux)

### Debugging
Download the [Mugen dependencies](https://github.com/Windblade-GR01/Ikemen_GO-Elecbyte-Screenpack) and unpack them into the Ikemen_GO source directory.  
Then, use [Goland](https://www.jetbrains.com/go/) or [Visual Studio Code](https://code.visualstudio.com/) to debug.

### Cross-compiling binaries with docker (Linux/Windows/MacOS)
The easiest way to compile binaries for other platforms is with Docker.  
You don't need the native development environment set to be able to build binaries if you decide to use Docker.  
The image downloaded has all required tools to compile Ikemen_GO for all the three platforms.

Install [docker for your platform](https://www.docker.com/get-started).  
For MacOS, you can install using homebrew (`brew cask install docker`).

Open a terminal, go to Ikemen `build` directory folder and then run the script `build_docker.sh`.  
Look inside the script for details on how it works.

### Preparing for release
Before generating the installation bundle, first make sure that the binaries for Ikemen_GO are properly generated.  
Download and install [InstallBuilder](https://installbuilder.bitrock.com).  
Once finished, open the program, then open the file releaseconf.xml.  
Click in Build.  
For other platforms, select the target platform then click in build.

You may edit releaseconf.xml or use the InstallBuilder wizard to customize the installer.

NOTE: InstallBuilder is free for opensource projects. But you need to [get a license for it](https://installbuilder.bitrock.com/open-source-licenses.html).  
Do not include copyrighted dependencies in the bundle.

## Features added since Mugen
Refer to the wiki article [Details of new features](https://github.com/Windblade-GR01/Ikemen_GO/wiki/Details-of-new-features) to see new features added that are not available in Mugen 1.1 and bellow.

## References
Suehiro repo (Original creator of the engine)  
https://osdn.net/users/supersuehiro/pf/ikemen_go/

Ikemen GO, K4thos fork of Ikemen. (Commonly updated and merged constantly to this repo)  
https://github.com/K4thos/Ikemen-GO-Plus

The default motif bundled with the engine:  
https://github.com/ikemen-engine/Ikemen_GO-Elecbyte-Screenpack

## What I.K.E.M.E.N means.
Ikemen is an acronym of:

**い**つまでも **完**成しない **永**遠に **未**完成 **エン**ジン  
**I**tsu made mo **K**ansei shinai **E**ien ni **M**ikansei **EN**gine

## Licence
The code is under the MIT Licence.  
Non-code assets are under CC-BY 3.0.

Check [License.txt](License.txt) for more deatils.
