So um this is a alternative build huide.

First you need these:
> ┬ MinGW                           (<https://sourceforge.net/projects/mingw-w64/files/>)
>
> └─ I use the SEH version called ``x86_64-posix-seh``
>
> ─ GO-lang                          (<https://golang.org/>)
>
> ─ Git                                    (<https://git-scm.com/>)
Git is required for Go-lang's ``get`` command to work. (Used for building)

After installing everithing but Before working on the code you need to run the ``get.cmd`` file inside the ``build`` folder.

Once you have opened the Ikemen GO project folder you need to edit the ``go.gopath`` inside VS code workspace settings (Not global setting) to point to the GO folder inside the project folder.

If you use the portable version you have to set up these environment variables:
> GOPATH = ``%USERPROFILE%\Go``
>
> GOROOT = ``<Go-lang installation directory>``
>
> Path =+ ``<Go installation directory>\bin``
>
> Path =+ ``<MinGW installation directory>\bin``
Path is a special variable that can contains multiple sub variables so you add them to to Path instead of creating it.

Also you need soft openAL libraries (<https://kcat.strangesoft.net/openal.html>)

Download ``openal-soft-1.19.1-bin.zip``

Inside the ZIP fie you will find a folder structure like this:

![IMG](https://media.discordapp.net/attachments/233363722934943744/631028770069020693/OpenAL_Soft_folder_structure.png)

Extract the AL ``folder`` (Purple box) inside MinGW's ``include`` folder.

Exact the contents **inside** the ``Win64/Win32`` (Depending of your OS version) inside MinGW ``lib`` folder.

**PS:** Depending of the MinGW version you also have to extract these files again to the folder duplicates in the ``x86_64-w64-mingw32`` folder.

Once you have compile the engine copy and rename ``soft_oal.dll`` to ``OpenAL32.dll`` and put it besides the ``Ikemen_GO.exe``
