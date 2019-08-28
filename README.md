# GoUp [![Travis-CI](https://travis-ci.com/worldiety/gomobilebuilder.svg?branch=master)](https://travis-ci.com/worldiety/gomobilebuilder) [![Go Report Card](https://goreportcard.com/badge/github.com/worldiety/goup)](https://goreportcard.com/report/github.com/worldiety/goup) [![GoDoc](https://godoc.org/github.com/worldiety/goup?status.svg)](http://godoc.org/github.com/worldiety/goup)  
GoUp (pronounced go-up) is an install and make tool which helps to build go modules with 
gomobile for android and ios. It contains an automatic versioned toolchain provisioning,
emulated module (vgo) support for gomobile and an build artifact cache.  

The motivation behind this tool is to bring the simplicity into the gomobile android world
again. Today there are a lot of pitfalls to create a working setup, so we
integrated all current workarounds for Linux/amd64 and MacOS/amd64 for a single selected
toolchain into the tool. The name is inspired from *rustup*, the rust toolchain installer. 
Like rustup, GoUp will manage different toolchains (Android NDK, SDK, Java JDK and Go) 
in different versions, even for distinct projects, to provide you a stable build experience.

What is still missing is extensive testing, configuration of more NDK, SDK, JDK and Go versions 
which are known to be compatible with each other, windows support and versioning of gomobile.
Also the example project layout is created to support a crossplatform project
layout, but still only contains an android and ios project.

Our perspective is to evaluate the practicality and to develop and support this tool in the 
case of success.

## Example

We've create an Android and iOS [example](https://github.com/worldiety/goup/tree/master/example) 
for your pleasure. It contains the android module *libGo* which invokes the GoUp wrapper
script (goupw) which in turn downloads the actual GoUp version for your current platform.
Please note, that the *goupw* script is only a crude bash script, were you need to define
the platforms to build (`TARGETS`) and which GoUp version to use (`VERSION`).

The gradle script in *libGo* builds the go library intentionally in every configuration phase,
which ensures that you have always the valid generated Java API at your fingertips. 
Performance wise the impact is reduced, due to the GoUp artifact caching, which avoids 
building if nothing has changed, so in subsequent configure calls the process completes within
a few milliseconds. 

To build, you just need to open the project with Android Studio and wait some minutes, until
GoUp has downloaded and installed all required toolchains. To get some 'entertainment', you
have to switch to the *Build Output* window and toggle the view in Android Studio.

In general you should not check in the generated artifacts, because the generated aar
file may not be deterministic (bitwise).

The process in iOS is very comparable and integrates the GoUp compilation phase simply as
a *script phase* in the *build phases* section (just after *Target Dependencies*). Note, that you also
need to disable the bitcode flag in *Build Settings*. The generated framework becomes fairly large, so
a check in may also be inadequate.

## how to build

The core of a GoUp project is the goup.yaml file, which contains all declared versions and related
build information to process your project.

The build yaml file may look like this:

```yaml
# The name is used to setup a custom workspace and tools.
# You should not invoke parallel builds for the same project
name: MySuperProject

# set custom environment variables, which are always applied into the executing environment, just
# like they have been defined before invoking GoUp
variables:
  TEST: "HELLO WORLD"
  TEST2: "HELLO WORLD"

# before_script is executing the following commands before the actual build starts. You can use it, to e.g. work around
# authentication problems with go get and git
before_script:
  - git config --global url."https://user:password@gitlab.mycompany.com/".insteadOf "https://gitlab.mycompany.com/"


# The build section defines what and how goup should work
build:
  # We want a gomobile build, e.g. for ios or android
  gomobile:

    # the toolchain section is required to setup a stable gomobile building experience
    toolchain:
      # which go version?
      go: 1.12.4
      # which android ndk version?
      ndk: r19c
      # which android sdk version?
      sdk: 4333796
      # which java version?
      jdk: 8u212b03
      # which gomobile version?
      gomobile: wdy-v0.0.1

    # The ios section defines how our iOS library is build. This only works on MacOS with XCode installed
    ios:
      # The gomobile -prefix flag
      prefix: MyLib

      # The gomobile -o flag, this will be a folder
      out: ./appIOS/MyLib.framework

      # The gomobile -bundleid flag sets the bundle ID to use with the app.
      bundleid:

      # The gomobile -ldflags flag
      ldflags:


    # The android section defines how our android build is executed
    android:
      # The gomobile -javapkg flag prefixes the generated packages
      javapkg: com.mycompany.myproject

      # The gomobile -o flag, this will be an android archive file. You should only ever use
      # a single go library in your app. Otherwise there may be some technical issues and
      # it also wastes a lot of storage and memory resources in your app.
      out: ./appAndroid/libGo/libs/fatLib.aar

      # The gomobile -ldflags flag
      ldflags:


    # The modules section defines a list of all local or remote go modules, which should be included in the build.
    # You can have more than one, but probably you only need a single one and want to use
    # real go mod dependencies instead.
    modules:
      - ./libGo/mycompany/myproject

    # The export section defines all exported packages which are passed to gobind by gomobile.
    # Gomobile does not generate transitives exports, so you need to declare all
    # packages containing types and methods which you want to have bindings for.
    # Be careful with name conflicts, because the last part of the package will be used
    # to scope the types.
    export:
      # contains handsome wrappers to allow passing unsupported types (interfaces, maps, slices)
      # through gomobile. In our case pkgb wants to export those types.
      - github.com/worldiety/std
      # our actual local packages
      - mycompany/myproject
      - mycompany/myproject/pkga
      - mycompany/myproject/pkgb



```

GoUp has a few optional commandline arguments, and also evaluates the environment variable GOUP_HOME.

```bash
goup -help
  -buildFile string
        Use a build file to load. (default "./goup.yaml")
  -clean
        Removes the project workspace, but keeps toolchains.
  -dir string
        Use a custom directory to resolve relative paths from goup.yml. 
  -help
        Shows this help
  -home string
        Use this as the home directory, where GoUp holds toolchains, projects and workspaces. 
  -loglevel int
        The LogLevel determines what is printed into the console. 0=Debug, 1=Info, 2=Warn, 3=Error
  -reset
        Performs a reset, delete the home directory and exits
  -resources string
        XML which describes downloadable toolchains (default "https://raw.githubusercontent.com/worldiety/goup/master/resources.xml")
  -targets string
        The targets to build, e.g. gomobile/android or gomobile/ios. Can be concated by : (default "all")
  -version
        Shows the version
```

You always need an *export* list and every exported module should be declared (at least transitively)
from your *module* projects. All referred dependencies are upgraded and copied into
an artificial go path in `~/.goup/<project>/go`, so that gomobile is happy. You can also
define more than one local module, e.g. if you have multiple local dependent or related
modules. But probably it would be better to only ever have a single local module and
refer to external versioned go dependencies.

Toolchains are installed in `~/.goup/toolchains`, one for each type and version. Also
GoUp uses interprocess filelocks for modifying toolchains and projects, to allow
at least concurrent (but sequentialized) builds without corruptions.


## hint
Important things like maps, slices, derived basic and value types won't work with gomobile. 
To mitigate these limitations, you can use the module https://github.com/worldiety/std. 
Do not forget to export it, as shown in the example.
