package main

import (
	"encoding/json"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/yuin/gopher-lua"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
)

func init() {
	runtime.LockOSThread()
}
func chk(err error) {
	if err != nil {
		panic(err)
	}
}
func main() {
	chk(glfw.Init())
	defer glfw.Terminate()
	defcfg := []byte(strings.Join(strings.Split(`{
  "HelperMax": 56,
  "PlayerProjectileMax": 50,
  "ExplodMax": 256,
  "AfterImageMax": 8,
  "Attack.LifeToPowerMul": 0.7,
  "GetHit.LifeToPowerMul": 0.6,
  "Width": 640,
  "Height": 480,
  "Super.TargetDefenceMul": 1.5,
  "LifebarFontScale": 0.5,
  "System": "script/main.lua",
  "KeyConfig": [
    {
      "Joystick": -1,
      "Buttons": ["UP", "DOWN", "LEFT", "RIGHT",
        "z", "x", "c", "a", "s", "d",
        "RETURN"]
    }
  ],
  "_comment": {
    "_comment": "ジョイスティック (0番) の場合の KeyConfig",
    "KeyConfig": [
      {
        "Joystick": 0,
        "Buttons": [-3, -4, -1, -2,
          1, 2, 7, 0, 3, 5,
          9]
      }
    ]
  }
}
`, "\n"), "\r\n"))
	tmp := struct {
		HelperMax              int32
		PlayerProjectileMax    int
		ExplodMax              int
		AfterImageMax          int32
		Attack_LifeToPowerMul  float32 `json:"Attack.LifeToPowerMul"`
		GetHit_LifeToPowerMul  float32 `json:"GetHit.LifeToPowerMul"`
		Width                  int32
		Height                 int32
		Super_TargetDefenceMul float32 `json:"Super.TargetDefenceMul"`
		LifebarFontScale       float32
		System                 string
		KeyConfig              []struct {
			Joystick int
			Buttons  []interface{}
		}
	}{}
	chk(json.Unmarshal(defcfg, &tmp))
	const configFile = "script/config.json"
	if bytes, err := ioutil.ReadFile(configFile); err != nil {
		f, err := os.Create(configFile)
		chk(err)
		f.Write(defcfg)
		chk(f.Close())
	} else {
		if len(bytes) >= 3 &&
			bytes[0] == 0xef && bytes[1] == 0xbb && bytes[2] == 0xbf {
			bytes = bytes[3:]
		}
		chk(json.Unmarshal(bytes, &tmp))
	}
	sys.helperMax = tmp.HelperMax
	sys.playerProjectileMax = tmp.PlayerProjectileMax
	sys.explodMax = tmp.ExplodMax
	sys.afterImageMax = tmp.AfterImageMax
	sys.attack_LifeToPowerMul = tmp.Attack_LifeToPowerMul
	sys.getHit_LifeToPowerMul = tmp.GetHit_LifeToPowerMul
	sys.super_TargetDefenceMul = tmp.Super_TargetDefenceMul
	sys.lifebarFontScale = tmp.LifebarFontScale
	stoki := func(key string) int {
		return int(StringToKey(key))
	}
	for _, kc := range tmp.KeyConfig {
		b := kc.Buttons
		if kc.Joystick >= 0 {
			sys.keyConfig = append(sys.keyConfig, KeyConfig{kc.Joystick,
				int(b[0].(float64)), int(b[1].(float64)),
				int(b[2].(float64)), int(b[3].(float64)),
				int(b[4].(float64)), int(b[5].(float64)), int(b[6].(float64)),
				int(b[7].(float64)), int(b[8].(float64)), int(b[9].(float64)),
				int(b[10].(float64))})
		} else {
			sys.keyConfig = append(sys.keyConfig, KeyConfig{kc.Joystick,
				stoki(b[0].(string)), stoki(b[1].(string)),
				stoki(b[2].(string)), stoki(b[3].(string)),
				stoki(b[4].(string)), stoki(b[5].(string)), stoki(b[6].(string)),
				stoki(b[7].(string)), stoki(b[8].(string)), stoki(b[9].(string)),
				stoki(b[10].(string))})
		}
	}
	l := sys.init(tmp.Width, tmp.Height)
	if err := l.DoFile(tmp.System); err != nil {
		switch err.(type) {
		case *lua.ApiError:
			errstr := strings.Split(err.Error(), "\n")[0]
			if len(errstr) < 10 || errstr[len(errstr)-10:] != "<game end>" {
				panic(err)
			}
		default:
			panic(err)
		}
	}
}
