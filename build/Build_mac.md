## Setup the development environment. (execute only once)
For mac, the easier way is using homebrew. Install homebrew [following these instructions](https://brew.sh)
Next, open a terminal and use homebrew to install the dependencies.

```
brew install caskroom/cask/brew-cask
brew install git go
brew install openal-soft
```

The following packages are not required but they makes it a lot easier to code. 
```
brew cask install goland visual-studio-code
```
Get the code:
```
git clone 
```
## Compiling
Open a terminal, move to downloaded folder:

`cd Ikemen_GO`

Then, move to the build folder:

`cd build`

Execute get.sh to download Ikemen dependencies (it takes a while):

`./get.sh`

FINALLY compile:

`./build.sh`

The compiled Ikemen GO binary now should be inside the bin folder.

And now, Ikemen can be opened double clicking Ikemen_GO.command, or with the terminal:

`./Ikemen_GO_mac`

PS: If you want to run the engine you can to donwload the mugen font and screenpack files at this [link](https://drive.google.com/file/d/1YCp5LYq-v1LSCFiq7a4RxLmJHdwZ4b0s/view),
they need to be extracted at the `bin` directory.