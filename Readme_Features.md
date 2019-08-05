# Details of additional functions

## About Map

So-called associative array. A character string and an integer value can be linked and set for each character. In addition to being used as a variable, you can refer to the opponent, so you can set an attribute name and use it as a tag, or use it as a flag for a specific technique. Note that spaces cannot be used in map names.

As an example, a map can be set by adding the following description to the character def file.
```ini
[Map]
Ryu = 1
Streetfighter = 1
man = 1
birthyear = 1964
Japan = 1
Ansatsuken = 1
```
Set an integer value with = to the name of the key map (string). If you have this map, from top to bottom

* Is Ryu
* Is from Street fighter
* Is a man
* Was born in 1964
* Is from Japan
* Uses Ansatsuken
* It can be recognized from the other party that it is a character

This map can be recognized by a trigger.
```ini
trigger1 = map (birthyear)
```
Use the name of the map you want to recognize in parentheses. For example, a character with the above map will return map (birthyear) as 1964. If nothing is set, 0 is returned.

#### State controller that changes Map value
```ini
[State Test]
type = mapset
trigger1 =! time
map = "birthyear"
value = 1987
```
```ini
map = "Key map name" (string)
value = number to enter into the map (integer)
```
A map can also be set in the state controller. It can be used to change a number that has already been set or to set a new map.

## Additional state controllers

### MatchRestart
````ini
[State Test]
type = MatchRestart
trigger1 = time = 10
p1def = "kfm.def"
p2def = "kfm.def"
Stagedef = "stage0.def"
reload = 1, 1
````
Reset the round and resume. (Same effect as F4 of debug key)

If Reload is set to 1, the file is reloaded and the round is restarted from the beginning. (Same effect as Shift + F4 debug key)

Reload 1 item specifies whether to reload P1 and 2 item whether P2 is reloaded. If all are 0, the round is reset without reloading.

If the path of the def file is specified with P1def, P2def, and Stagedef, the file is read when reloading. The path at that time is based on the execution folder character.

### MapSet
See above.

### Savefile
```ini
[State Test]
type = SaveFile
trigger1 = time = 10
savedata = var
path = "kfm.gob"
```
Put specified data together and save it as binary. It uses gob, which is a serialized format for Go language, as the storage format.

Specify the data saved by savedata as var, fvar, or map. All characters specified by the character or helper who executed the function are stored at that time.

Specify the save destination file path by path (execution character standard). Can use any extension (.gob is recomneded)

### Loadfile
```ini
[State Test]
type = LoadFile
trigger1 = time = 10
savedata = var
path = "kfm.gob"
```
Loads the specified data and overrides the data of the execution character. Note that all the data before reading will disappear.

Specify the data to be read by savedata from var, fvar, or map. An error occurs if you make a mistake in the path.

Specify the path of the file to be read in path (Relative to the character folder)

## Additional triggers

### Majorversion
Returns 1 if mugenversion in the def file is 1.0 or higher.

### Map
See above.

### Selfstatenoexist
```ini
trigger1 = Selfstatenoexist(3000)
```
Returns 1 if there is a Statedef with the specified number. Otherwise it returns 0. Use the Statedef number you want to recognize in parentheses.

### Stagebackedge
Returns the distance to the stage edge behind you.

### Stagefrontedge
Returns the distance to the stage edge in front of you.

## Changes to triggers

### TeamMode
Now Teammode can also return "Tag"
```ini
trigger1 = TeamMode = Tag
```

# Details of additional parameters
## Additional parameters for the state controller

### Roundnotskip
Parameter specified in the Assertspecial flag. During execution, you can no longer skip the intro and victory pose by pressing a button.

### Teamside
Hitdef parameters. Hitdef be treated as an attack from the Teamside you specify (similar to the trigger of Teamside).

### Extendsmap
Helper parameters. When set to 1, the parent map is inherited by Helper.

### Projangle
Projectile parameters. Specifies the angle to rotate the Projectile animation.

### ReadplayerID
Selftate parameters. Change to the state of the character with the specified Player ID. If successful, it would take the character with the specified PlayerID to the selected state.

### RedirectID
Can be used for all state controllers. An optional parameter that changes the executor of the statedef to a specified PlayerID character.
It is possible to interfere with the target without taking the target.
You can easily implement the behavior of a so-called persistent target.
The main use is to reproduce poisons that reduce life without touching the opponent.

The following stacons do not operate for the convenience of processing that is effective only in the execution frame.
* Posfreeze
* Trans
* Part of Assertspecial (Screenbound, Playerpush, Angledraw, Offset effect exits when executed after the target ID)

## Additional parameters for StageDef

### Attachedchar
Write to Info in stage def. The character with the specified path appears as the character on the stage side.
Teamside hits 3 and the attack hits both the 1P side and the 2P side.
The main use is to allow characters to realize functions that cannot be implemented with stage def alone.

### Autoresizeparallax
BG layer parameters. Specifies whether to automatically correct the parallax size by IKEMEN when zooming out. 1 if omitted.
By specifying Zoomdelta after setting this to 0, you can reproduce the same behavior as when parallax zoomed out with MUGEN 1.1.

### Zoomscaledelta
BG layer parameters. Zoomdelta can be reverse-corrected. The main use is to reverse the correction of the image ratio when zooming out.
It is also possible to read two values, x and y, and omit each and apply reverse correction to only one of them.

### Xbottomzoomdelta
BG layer parameters. Specifies the X scale reduction ratio of the bottom of the image when zooming out. 1 if omitted.
Does not work unless Autoresizeparallax is 1.
Measures against the phenomenon where the X scale on the bottom becomes abnormally small and freezes when it is severely zoomed out on a stage using 3D parallax.
Since this phenomenon is a correct operation in the specification, a new parameter has been added to enable countermeasures by description.

## Additional parameters for character def file

### Portraitscale
Specify the display magnification of the portrait to be written to Info (Localcoord-compliant display magnification is overwritten)

# Misc. Info

### About Zoomdelta on stage
Two items of Zoomdelta are not read by MUGEN1.1, but here they are processed as Zoomdelta in the Y direction.
If omitted, the same value as 1 item is entered and the same processing as MUGEN1.1 is performed.

### About BGctrl SinY and SinX
MUGEN1.1 has a bug that BGctrl's SinY and SinX ignore the looptime of BGCtrlDef.
Therefore, in IKEMEN GO, the movement of SinY and SinX on the stage assuming MUGEN 1.1 may be strange.
This can be dealt with by combining the BGCtrlDef looptime and the SinY or SinX Value 2 items (cycle).

### About Explod Xangle and Yangle
In MUGEN1.1, if this is used, the viewing angle works slightly, so the image is arbitrarily distorted depending on the position,
but IKEMEN GO does not have this distortion because it is drawn in orthographic projection.
At the moment, we can't confirm the character that makes the effect strange by eliminating this distortion,
and I think that it is better to use this specification because I think it is easier for the producer to make it without distortion.

### Reasons for not handling boundhigh fluctuation processing by Zoomout
In MUGEN 1.1, boundright and boundleft fluctuate depending on the value of Zoomout, so IkemenGO also has a corresponding process.

Similarly, the boundhigh also changes depending on the zoomout value.
However, in the case of the boundhigh, the correction process is not performed because the value to be corrected is not constant depending on the stage.
The reason why this value to be corrected is not constant is because the zoomdelta of the background layer that determines
the boundbound high of the stage differs depending on the stage.
In boundright and boundleft, there is almost always a background layer with a reference zoomdelta of 1,
so the value that shifts due to the zoomout is constant regardless of the stage.
On the other hand, boundhigh has a zoomdelta of the layer corresponding to the upper limit of the background that determines it
 0.2 if it does not scroll so much, or 1 in the case of a single picture stage.
In this case, the correction value is smaller in the stage where the zoomdelta is 0.2 than in the stage where the zoomdelta is 1.
And just read def to determine which layer determines the boundhigh (zoomdelta is not necessarily the smallest layer).
Therefore, the boundhigh gap cannot be corrected automatically.
This is not a problem in MUGEN 1.1 because there is a bug that will not reach the stage display limit unless zoomed out maximum.
And since boundhigh may be set according to this, the value to be corrected by the stage is not constant.
There are also a number of producers who ignore the 1.1 zoom-out bug and set boundhigh.
In this case, no correction is performed as intended by the autor.
