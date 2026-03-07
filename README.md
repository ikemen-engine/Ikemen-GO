# Ikemen GO

Ikemen GO is an open source fighting game engine that supports resources from the [M.U.G.E.N](https://en.wikipedia.org/wiki/Mugen_(game_engine)) engine, written in Google’s programming language, [Go](https://go.dev/). It is a complete rewrite of a prior engine known simply as Ikemen.

## Features
Ikemen GO aims for backwards-compatibility on par with M.U.G.E.N version 1.1 Beta, while simultaneously expanding on its features in a variety of ways.

Refer to [our wiki](https://github.com/ikemen-engine/Ikemen-GO/wiki) to see a comprehensive list of new features that have been added in Ikemen GO.

## Installing
Ready to use builds for Windows, macOS and Linux can be found in the [releases section](https://github.com/ikemen-engine/Ikemen-GO/releases) of this repository. You can find nightly builds [here](https://github.com/ikemen-engine/Ikemen-GO/releases/tag/nightly) as well, which update on every commit.

## Running
Download the ZIP archive that matches your operating system and extract its contents to your preferred location.

On Windows, double-click `Ikemen_GO.exe`.
On macOS or Linux, double-click `Ikemen_GO.command`.

## Developing
These instructions are for those interested in developing the Ikemen GO engine itself. Instructions for creating custom stages, fonts, characters and other resources can be found in the community forum.

### Building
For setup and platform-specific steps, see [BUILDING.md](./BUILDING.md).
It covers Windows, Linux (including ARM64), macOS (Apple Silicon and Intel), and Android (APK via Docker).

### Debugging
In order to run the compiled Ikemen GO executable, you will need to download the [engine dependencies](https://github.com/ikemen-engine/Ikemen-GO-Screenpack) and unpack them into the Ikemen-GO source directory. After that, you can use [Goland](https://www.jetbrains.com/go/) or [Visual Studio Code](https://code.visualstudio.com/) to debug.

## Troubleshooting
If you run into any issues with Ikemen Go, you can report it on our [issue tracker](https://github.com/ikemen-engine/Ikemen-GO/issues). It is recommend to read [this page](https://github.com/ikemen-engine/Ikemen-GO/blob/develop/CONTRIBUTING.md) before submitting a bug report.

## References
- [The original reposity of Ikemen GO.](https://osdn.net/users/supersuehiro/pf/ikemen_go/) This project was forked from this repository due to its original author seemingly abandoning the project.

- [The default motif bundled with the engine.](https://github.com/ikemen-engine/Ikemen-GO-Screenpack) Note that this motif is licensed under CC-BY 3 rather than Ikemen GO's source, which is MIT.

## Name
"Ikemen" is an acronym of:

**い**つまでも **完**成しない **永**遠に **未**完成 **エン**ジン  
**I**tsu made mo **K**ansei shinai **E**ien ni **M**ikansei **EN**gine

## License
Ikemen GO engine is under the MIT License.
Bundled screenpack assets are under Creative Commons licenses.
See [LICENSE.txt](LICENSE.txt) for more details.
This program statically links FFmpeg (LGPL v2.1).

The exact corresponding source for the FFmpeg build is provided on the [release page](https://github.com/ikemen-engine/Ikemen-GO/releases/latest) as Source-code-FFmpeg.tar.gz. You may rebuild this application against a modified FFmpeg.
