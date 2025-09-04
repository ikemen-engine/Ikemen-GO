; Command File Template for Ikemen GO
; Each parameter references M.U.G.E.N documentation and notes compatibility.

[Remap]
; [MUGEN-Compat] Button remapping (M.U.G.E.N docs: CMD -> Remap)
; Example: a = x

[Defaults]
command.time = 15 ; [MUGEN-Compat] Frames allowed to enter command
buffer.time = 1 ; [Ikemen-Only] Input buffer frames

; [EXAMPLE] Quarter circle forward punch
[Command]
name = "qcf_x" ; [MUGEN-Compat] Command name
command = ~D, DF, F, x ; [MUGEN-Compat] Input sequence
time = 20 ; [MUGEN-Compat] Frames to complete command

; [EXAMPLE] 360 motion
[Command]
name = "spin" ; [MUGEN-Compat]
command = ~F, DF, D, DB, B, UB, U, UF ; [MUGEN-Compat]
time = 30 ; [MUGEN-Compat]

[State -1]
; Example of triggering a command
[State -1, QCF]
type = ChangeState ; [MUGEN-Compat]
value = 3000 ; [MUGEN-Compat]
triggerall = command = "qcf_x" ; [MUGEN-Compat]
trigger1 = statetype = S ; [MUGEN-Compat]
