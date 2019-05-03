# gomobilebuilder: Android example

We will not use the gradle plugin of gomobile, because it has been deprecated and is likely to be removed in the
future (see https://github.com/golang/go/issues/25314).
 However that is not a problem, because gomobilebuilder and a simple gradle module can replace that properly. 

## preface
You need to install go and configure the go path bin folder to be on your path.

```bash
# e.g. for macos, go to https://golang.org/ and install the pkg installer
go version # should print something like: go version go1.12 darwin/amd64
 
# edit your current bash paths
nano ~/.bash_profile

# add/change the following lines
export PATH=${PATH}:~/go/bin/:/Applications/android-sdk/platform-tools:/Applications/android-ndk-r10
export ANDROID_HOME=/Applications/android-sdk
export ANDROID_NDK_HOME=/Applications/android-ndk-r19c
export ANDROID_SDK_ROOT=/Applications/android-sdk
  
# load the new config
source ~/.bash_profile
```

## gradle
Setup your Android project as usual. Now we will create a new gradle module, which will provide the gomobile
aar file and the invocation of gomobilebuilder.

```bash
# go into your android application project, this directory contains at least an app folder, build.gradle and settings.gradle
cd myproject

# setup a new module
mkdir -p fatlib/src/main/go/fatlib
touch fatlib/gmb.sh
touch fatlib/gomobile.build
touch fatlib/build.gradle
touch fatlib/src/main/go/fatlib/go.mod
touch fatlib/src/main/go/fatlib/go.main

# register new module by adding your gradle module, e.g.: include ':app', ':fatlib'
nano settings.gradle

```

Content of fatlib/gmb.sh
```bash
#!/usr/bin/env bash

# this script bootstraps the gomobilebuilder setup

set -e
export GO111MODULE=on

if ! [ -x "$(command -v go)" ]; then
  echo 'Error: go is not installed.' >&2
  exit 1
fi


if ! [ -x "$(command -v gomobilebuilder version)" ]; then
  go install -u github.com/worldiety/gomobilebuilder
fi

gomobilebuilder -version
gomobilebuilder
```

Content of gomobile.build
```json
{
  "project":"fatlib",
  "build":{
    "android":{
      "pkg": "de.worldiety.example",
      "out": "./fatlib.aar"
    }
  },
  "import":[
    "./src/main/go/fatlib"
  ],
  "export":[
    "worldiety/example/fatlib",
    "github.com/worldiety/std"
  ]
}
```

Content of build.gradle
```gradle
configurations.maybeCreate("default")
configurations.maybeCreate("source")
artifacts.add("default", file('fatlib.aar'))
artifacts.add("source", file('fatlib-sources.jar'))


def proc = 'sh ./gmb.sh'.execute(null, new File("fatlib"))
proc.waitForProcessOutput(System.out, System.err)
```

Content of go.mod
```go
module worldiety/example/fatlib

go 1.12

require (
    github.com/worldiety/std v0.0.0-20190429141453-4964c97755c6 // indirect
    github.com/worldiety/xobj v0.0.0-20190426163538-01ff3dba5c17
)
```

Content of main.go
```go
package fatlib

import (
	"github.com/worldiety/std"
	)

// HelloWorld says hello
func HelloWorld() string {
    return "Hello World"
}

func NewBox() *std.Box {
	return &std.Box{}
}

type Callback interface{
    DoStuff()
}

func InvokeMe(callback Callback){
    callback.DoStuff()
}

func main() {

}

```

Last, insert in your app/build.gradle the dependency

```gradle
dependencies {
    //...
     implementation project(":fatlib")
}
```