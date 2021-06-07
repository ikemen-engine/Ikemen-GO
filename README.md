# Ikemen GO
IKEMEN Go is a remake of the IKEMEN (open source fighting games engine that supports M.U.G.E.N resources) in Google’s Programming Language “Go”.

## Installing
For now automated build from AppVeyor can be found at:  
https://ci.appveyor.com/project/Windblade-GR01/ikemen-go  
Inside the artifacts tab. Installation bundles will be provided in the future.

If you need to install OpenAL dependencies, for windows, look at https://kcat.strangesoft.net/openal.html.  
For other platforms, use the respective package manager.

## Running
On windows, execute `Ikemen_GO.exe`

On mac/linux, double-click on `Ikemen_GO.command`

## Developing
These instructions are for those interested in developing the Ikemen_GO engine. Instructions on contributing with custom stages, fonts, characters and other resources can be found in the community forum.

### Building on Windows
Check the instructions [here](https://github.com/Windblade-GR01/Ikemen_GO/wiki/Building-on-Windows)

### Building on Mac
Check the insturctions [here](https://github.com/Windblade-GR01/Ikemen_GO/wiki/Building-on-MacOS)

### Building on Linux
Check the instructions [here](https://github.com/Windblade-GR01/Ikemen_GO/wiki/Building-on-Linux)

### Debugging
Download the [Mugen dependencies](https://github.com/Windblade-GR01/Ikemen_GO-Elecbyte-Screenpack) and unpack them into the Ikemen_GO source directory. Then, use [Goland](https://www.jetbrains.com/go/) or [Visual Studio Code](https://code.visualstudio.com/) to debug.

### Cross-compiling binaries with docker (linux/windows/mac)
The easiest way to compile binaries for other platforms is with docker.  
You don't need the native development environment set to be able to build binaries if you decide to use docker.  
The image downloaded has all required tools to compile Ikemen_GO for all the three platforms.

Install [docker for your platform](https://www.docker.com/get-started). For mac, you can install using homebrew (`brew cask install docker`).

Open a terminal, go to Ikemen source directory folder and then run the script build_docker.sh  
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

Ikemen GO Plus, K4thos fork of Ikemen (Some features were borrowed from his repo)  
https://github.com/K4thos/Ikemen-GO-Plus

## What I.K.E.M.E.N means.
Ikemen is an acronym of:  
**I**tsu made mo **K**ansei shinai **E**ien ni **M**ikansei **EN**gine  
**い**つまでも **完**成しない **永**遠に **未**完成 **エン**ジン

## Licence
[MIT Licence](License.txt)

