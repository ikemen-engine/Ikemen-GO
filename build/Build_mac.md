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

Execute get.sh to download Ikemen dependencies (it takes a while):

`./get.sh`

FINALLY compile:

`./build.sh`

And now, Ikemen can be opened double clicking Ikemen-GO-Plus, or with the terminal:

`./Ikemen_GO.command`