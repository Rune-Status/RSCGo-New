/*
 * Copyright (c) 2019 Zachariah Knight <aeros.storkpk@gmail.com>
 *
 * Permission to use, copy, modify, and/or distribute this software for any purpose with or without fee is hereby granted, provided that the above copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 *
 */

package script

import (
	"context"
	"github.com/spkaeros/rscgo/pkg/server/log"
	"github.com/spkaeros/rscgo/pkg/server/world"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

var Scripts []string

var EngineChannel = make(chan func(), 20)
var InvTriggers []func(context.Context, reflect.Value, reflect.Value) (reflect.Value, reflect.Value)
var ObjectTriggers []func(context.Context, reflect.Value, reflect.Value) (reflect.Value, reflect.Value)
var BoundaryTriggers []func(context.Context, reflect.Value, reflect.Value) (reflect.Value, reflect.Value)
//var NpcTriggers []func(context.Context, reflect.Value, reflect.Value) (reflect.Value, reflect.Value)
var LoginTriggers []func(player *world.Player)
var InvOnBoundaryTriggers []func(player *world.Player, object *world.Object, item *world.Item) bool
var NpcTriggers = make(map[int]func(*world.Player, *world.NPC))

func Run(fnName string, player *world.Player, argName string, arg interface{}) bool {
	env := WorldModule()
	err := env.Define("client", player)
	if err != nil {
		log.Info.Println("Error initializing scripting environment:", err)
		return false
	}
	err = env.Define("player", player)
	if err != nil {
		log.Info.Println("Error initializing scripting environment:", err)
		return false
	}
	err = env.Define(argName, arg)
	if err != nil {
		log.Info.Println("Error initializing scripting environment:", err)
		return false
	}
	for _, s := range Scripts {
		if !strings.Contains(s, fnName) {
			continue
		}
		stopPipeline, err := env.Execute(s +
			`
` + fnName + `()`)
		if err != nil {
			log.Info.Println("Unrecognized Anko error when attempting to execute the script pipeline:", err)
			continue
		}
		if stopPipeline, ok := stopPipeline.(bool); ok && stopPipeline {
			return true
		} else if !ok {
			log.Info.Println("Unexpected return result from an executed Anko script:", err)
		}
	}
	return false
}

func Clear() {
	InvTriggers = InvTriggers[:0]
	BoundaryTriggers = BoundaryTriggers[:0]
	ObjectTriggers = ObjectTriggers[:0]
	NpcTriggers = make(map[int]func(*world.Player, *world.NPC))
	LoginTriggers = LoginTriggers[:0]
	InvOnBoundaryTriggers = InvOnBoundaryTriggers[:0]
}

//Load Loads all of the scripts in ./scripts and stores them in the Scripts slice.
func Load() {
	err := filepath.Walk("./scripts", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Info.Println(err)
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, "ank") {
			env := WorldModule()
			_, err := env.Execute(load(path))
			if err != nil {
				log.Info.Println("Unrecognized Anko error when attempting to execute the script pipeline:", err)
				return nil
			}
			fn, err := env.Get("invAction")
			action, ok := fn.(func(context.Context, reflect.Value, reflect.Value) (reflect.Value, reflect.Value))
			if ok {
				InvTriggers = append(InvTriggers, action)
			}
			fn, err = env.Get("objectAction")
			action, ok = fn.(func(context.Context, reflect.Value, reflect.Value) (reflect.Value, reflect.Value))
			if ok {
				ObjectTriggers = append(ObjectTriggers, action)
			}
			fn, err = env.Get("boundaryAction")
			action, ok = fn.(func(context.Context, reflect.Value, reflect.Value) (reflect.Value, reflect.Value))
			if ok {
				BoundaryTriggers = append(BoundaryTriggers, action)
			}
		}
		return nil
	})
	if err != nil {
		log.Info.Println(err)
		return
	}
}

func load(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		log.Warning.Println("Error opening script file for object action:", err)
		return ""
	}
	data, err := ioutil.ReadAll(file)
	if err != nil {
		log.Warning.Println("Error reading script file for object action:", err)
		return ""
	}

	return string(data)
}
