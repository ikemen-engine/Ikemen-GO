[Defaults]
command.time = 15
command.buffer.time = 1
;---キー２回連続入力------------------------------------

[Command]
name = "FF"       
command = F, F
time = 20

[Command]
name = "BB"       
command = B, B
time = 20

;---受け身--------------------------------------------

[Command]
name = "recovery" 
command = y
time = 1

;---方向キー＋ボタン------------------------------------

[Command]
name = "down_a"
command = /$D,a
time = 1

[Command]
name = "down_b"
command = /$D,b
time = 1

[Command]
name = "down_c"
command = /$D,c
time = 1

;---ボタン単発------------------------------------------

[Command]
name = "a"
command = a
time = 1

[Command]
name = "b"
command = b
time = 1

[Command]
name = "c"
command = c
time = 1

[Command]
name = "x"
command = x
time = 1

[Command]
name = "y"
command = y
time = 1

[Command]
name = "z"
command = z
time = 1

[Command]
name = "up"
command = U
time = 1

[Command]
name = "down"
command = D
time = 1

[Command]
name = "fwd"
command = F
time = 1

[Command]
name = "back"
command = B
time = 1

[Command]
name = "start"
command = s
time = 1

[Command]
name = "hold_x"
command = /x

[Command]
name = "hold_z"
command = /z


;---方向キー |---------------------------------------
[Command]
name = "holdfwd"   
command = /$F
time = 1

[Command]
name = "holdback"  
command = /$B
time = 1

[Command]
name = "holdup"    
command = /$U
time = 1

[Command]
name = "holddown"  
command = /$D
time = 1

[Statedef -1]
;移動------------------------------------------------------------
[State -1, ダッシュ]
type = ChangeState
triggerall = command = "FF"
trigger1 = (StateType != A) && (Ctrl)
value = 100

[State -1, バックステップ]
type = ChangeState
triggerall = command = "BB"
trigger1 = (StateType != A) && (Ctrl)
value = 105

[State -1, 起き上がり]
type = ChangeState
triggerall = !var(35)
trigger1 = var(30) >= var(29)
trigger1 = stateno = 5110
value = 5120

[State -1, 移動起き上がり（前）]
type = ChangeState
triggerall = var(35) = 1
trigger1 = var(30) >= var(29)
trigger1 = stateno = 5110
value = 1000

[State -1, 移動起き上がり（後ろ）]
type = ChangeState
triggerall = var(35) = 2
trigger1 = var(30) >= var(29)
trigger1 = stateno = 5110
value = 1001

;オートガード--------------------------------------------------------
[State -1,ガード]
type = ChangeState
value = 120
triggerall = var(8)
triggerall = var(27) = 2
triggerall = StateNo!=[120,155]
triggerall = Ctrl||stateno = 21
trigger1 = inguarddist

[State -1,ガード]
type = ChangeState
value = 120
triggerall = var(27) = 1 || var(27) = 3
triggerall = StateNo!=[120,155]
triggerall = Ctrl||stateno = 21
trigger1 = inguarddist

;---移動その他-----------------------------------------------------------------------------
[State -1, アドガ立ち]
type = ChangeState
value = 2000
triggerall = var(27) = 3
trigger1 = stateno = [150,151]

[State -1, アドガ屈]
type = ChangeState
value = 2001
triggerall = var(27) = 3
trigger1 = stateno = [152,153]

[State -1, アドガ空中]
type = ChangeState
value = 2002
triggerall = var(27) = 3
trigger1 = stateno = 154
trigger2 = stateno = 155 && time <= 10

;基本行動------------------------------------------------------------
[State -1, 頭の位置チェック！]
type = ChangeState
triggerall = command = "a"
trigger1 = (StateType != A) && (Ctrl)
value = 200

[State -1, 体の位置チェック！]
type = ChangeState
triggerall = command = "b"
trigger1 = (StateType != A) && (Ctrl)
value = 201

[State -1, 射程距離チェック！]
type = ChangeState
triggerall = command = "c"
trigger1 = (StateType != A) && (Ctrl)
value = 300

[State -1, カウントダウンアタック1]
type = ChangeState
triggerall = command = "x"
triggerall = command != "holdfwd" 
triggerall = command != "holdback" 
trigger1 = (StateType = S) && (Ctrl)
value = 400

[State -1, カウントダウンアタック2]
type = ChangeState
triggerall = command = "x"
triggerall = command != "holdfwd" 
triggerall = command != "holdback" 
trigger1 = (StateType = C) && (Ctrl)
value = 405

[State -1, カウントダウンアタック3]
type = ChangeState
triggerall = command = "x"
trigger1 = (StateType = A) && (Ctrl)
value = 406

[State -1, 飛び道具P]
type = ChangeState
triggerall = command = "x"
triggerall = command = "holdfwd" 
trigger1 = (StateType != A) && (Ctrl)
value = 410

[State -1, 飛び道具H]
type = ChangeState
triggerall = command = "x"
triggerall = command = "holdback" 
trigger1 = (StateType != A) && (Ctrl)
value = 420

[State -1, パラメータ表示切替]
type = varadd
trigger1 = command = "y"
var(17) = 1

[State -1, ルーラー表示切替]
type = varadd
trigger1 = command = "z"
var(21) = 1

;特殊行動----------------------------------------------------
[State -1, 距離を取るよ]
type = ChangeState
triggerall = var(25)%4 != 0
triggerall = var(26)
trigger1 = (StateType != A) && (Ctrl)
value = 21

[State -1, しゃがんで待つよ]
type = ChangeState
triggerall = var(24)%3 = 1
triggerall = var(26) = 0
trigger1 = (StateType != A) && (Ctrl)
value = 11

[State -1, 跳んで待つよ]
type = ChangeState
triggerall = var(24)%3 = 2
trigger1 = (StateType != A) && (Ctrl)
trigger2 = stateno = 21
value = 40
