; State File Template for Ikemen GO
; Each parameter references M.U.G.E.N documentation and notes compatibility.

[Statedef 0]
type = S ; [MUGEN-Compat] State type: S/A/C/L
physics = S ; [MUGEN-Compat] Physics handling
movetype = I ; [MUGEN-Compat] Move type (I=Idle)

[State 0, ChangeAnim]
type = ChangeAnim ; [MUGEN-Compat] Change animation at state start
trigger1 = Time = 0 ; [MUGEN-Compat] Trigger on first frame
value = 0 ; [MUGEN-Compat] Animation number

[State 0, VelSet]
type = VelSet ; [MUGEN-Compat] Set velocity
trigger1 = 1 ; [MUGEN-Compat] Always active
x = 0 ; [MUGEN-Compat] X velocity
y = 0 ; [MUGEN-Compat] Y velocity

; [EXAMPLE] Ikemen-only afterimage effect
[State 0, AfterImage]
type = AfterImage ; [Ikemen-Only] Create afterimage effect
trigger1 = command = "qcf_x" ; [EXAMPLE] Use command from CMD
palcolor = 256 ; [Ikemen-Only]
palbright = 0,0,0 ; [Ikemen-Only]
