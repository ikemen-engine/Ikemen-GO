;===============================================================================
; File: common.cmd
; Purpose: Shared command definitions for all characters
; Safe-edit: Back up before modification; maintain syntax to avoid runtime errors
;===============================================================================
; Common Commands
; The commands defined in this file will be appended to every character's command list.
; Commands used by common states should be placed here to prevent crashing due to missing commands.
; Most of these are kept blank so that they won't interfere with a character trying to remap them.

; Trigger: defined when command input detected; Range: standard buttons; [PITFALL]: remapping may break shared states
[Command]
name = "recovery"
command = 
time = 1
buffer.time = 1
buffer.hitpause = 1
buffer.pauseend = 1

; Trigger: defined when command input detected; Range: standard buttons; [PITFALL]: remapping may break shared states
[Command]
name = "TagShiftBack"
command = d
time = 1
buffer.time = 1
buffer.hitpause = 1
buffer.pauseend = 1

; Trigger: defined when command input detected; Range: standard buttons; [PITFALL]: remapping may break shared states
[Command]
name = "TagShiftFwd"
command = w
time = 1
buffer.time = 1
buffer.hitpause = 1
buffer.pauseend = 1

; Trigger: defined when command input detected; Range: standard buttons; [PITFALL]: remapping may break shared states
[Command]
name = "holdfwd"
command =
time = 1
buffer.time = 1
buffer.hitpause = 1
buffer.pauseend = 1

; Trigger: defined when command input detected; Range: standard buttons; [PITFALL]: remapping may break shared states
[Command]
name = "holdback"
command =
time = 1
buffer.time = 1
buffer.hitpause = 1
buffer.pauseend = 1

; Trigger: defined when command input detected; Range: standard buttons; [PITFALL]: remapping may break shared states
[Command]
name = "holdup"
command =
time = 1
buffer.time = 1
buffer.hitpause = 1
buffer.pauseend = 1

; Trigger: defined when command input detected; Range: standard buttons; [PITFALL]: remapping may break shared states
[Command]
name = "holddown"
command =
time = 1
buffer.time = 1
buffer.hitpause = 1
buffer.pauseend = 1

; Every common command below this point is deprecated.
; It is recommended that a character that doesn't need commands for itself but still wishes to use the command trigger should still define the necessary commands in its own command file.

; Trigger: defined when command input detected; Range: standard buttons; [PITFALL]: remapping may break shared states
[Command]
name = "x"
command =
time = 1
buffer.time = 1
buffer.hitpause = 1
buffer.pauseend = 1

; Trigger: defined when command input detected; Range: standard buttons; [PITFALL]: remapping may break shared states
[Command]
name = "y"
command =
time = 1
buffer.time = 1
buffer.hitpause = 1
buffer.pauseend = 1

; Trigger: defined when command input detected; Range: standard buttons; [PITFALL]: remapping may break shared states
[Command]
name = "z"
command =
time = 1
buffer.time = 1
buffer.hitpause = 1
buffer.pauseend = 1

; Trigger: defined when command input detected; Range: standard buttons; [PITFALL]: remapping may break shared states
[Command]
name = "a"
command =
time = 1
buffer.time = 1
buffer.hitpause = 1
buffer.pauseend = 1

; Trigger: defined when command input detected; Range: standard buttons; [PITFALL]: remapping may break shared states
[Command]
name = "b"
command =
time = 1
buffer.time = 1
buffer.hitpause = 1
buffer.pauseend = 1

; Trigger: defined when command input detected; Range: standard buttons; [PITFALL]: remapping may break shared states
[Command]
name = "c"
command =
time = 1
buffer.time = 1
buffer.hitpause = 1
buffer.pauseend = 1

; Trigger: defined when command input detected; Range: standard buttons; [PITFALL]: remapping may break shared states
[Command]
name = "start"
command =
time = 1
buffer.time = 1
buffer.hitpause = 1
buffer.pauseend = 1

; Trigger: defined when command input detected; Range: standard buttons; [PITFALL]: remapping may break shared states
[Command]
name = "d"
command =
time = 1
buffer.time = 1
buffer.hitpause = 1
buffer.pauseend = 1

; Trigger: defined when command input detected; Range: standard buttons; [PITFALL]: remapping may break shared states
[Command]
name = "w"
command =
time = 1
buffer.time = 1
buffer.hitpause = 1
buffer.pauseend = 1

; Trigger: defined when command input detected; Range: standard buttons; [PITFALL]: remapping may break shared states
[Command]
name = "m"
command =
time = 1
buffer.time = 1
buffer.hitpause = 1
buffer.pauseend = 1

; Trigger: defined when command input detected; Range: standard buttons; [PITFALL]: remapping may break shared states
[Command]
name = "menu"
command =
time = 1
buffer.time = 1
buffer.hitpause = 1
buffer.pauseend = 1
