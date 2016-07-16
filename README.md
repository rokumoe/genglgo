# genglgo
generate gl.go for gl.xml

a better tool see the [glow](https://github.com/go-gl/glow)

## usage
1. get
```
go get github.com/vizee/genglgo
```

2. generate
```
cd $GOPATH/github.com/vizee/genglgo
./genglgo
```

3. use in glx
```go
package main

import (
    "gl"
    "x11"
    "x11/glx"
)

func main() {
    // 0. create window
    win := x11.CreateWindow("demo", 100, 100, 400, 400)
    x11.MapWindow(win)
    ctx := glx.CreateContext(win)
    // 1. make current context
    ctx.MakeCurrent();
    // 2. call gl.Init()
    gl.Init()
    win.OnExpose = func() {
        // 3. use gl API
        gl.ClearColor(1, 1, 0, 1);
        gl.Clear(gl.COLOR_BUFFER_BIT)
        glx.SwapBuffer(win)
    }
}
```
