# gomobilebuilder
The gomobilebuilder is a little helper to work around building issues with gomobile, like missing module support etc.

Sadly we need it: gomobilebuilder will pick up your local go modules, collects all dependencies, merges everything into an artifical
gopath, invokes gomobile and returns a fat crossplatform library back to you.

## installation

```bash
go get github.com/worldiety/gomobilebuilder
```

## Example

See the [Android tutorial](howto-android.md)

## how to build

Invoke the gomobilebuilder from within a directory of your choice. It must contain a valid gomobile.build file which may look like this
```json
{
  "project":"MyFatLib",
  "build":{
    "ios":{
      "prefix":"MyLib",
      "out":"./builds/MyLib.framework"
    },
    "android":{
      "pkg": "com.mycompany.myproject",
      "out": "fatlib.aar"
    }
  },
  "import":[
    "/avoid/absolute/paths/not/working/in/ci",
    "../../use/relative/paths/instead",
    "./my/local/module"
  ],
  "export":[
    "mydomain.tld/company/prj/pkg",
    "github.com/worldiety/std",
    "my/local/module"
  ]
}
```
The *build* section defines what to build, so either android or ios or both, each with their own parameters. 
The *import* section refers to local directories, which may be scattered throughout your developer (or ci) 
system and must be local go modules (with go.mod file). These modules and their dependencies are merged back 
into an artifical gopath, so that *gomobile* is happy and can find everything it needs. 
The gopath is located at `~/.gomobilebuilder/<project>/workspace`. The last section *export* 
specifies all packages from the merged gopath which you want to access through gomobile in your 
final library. Please keep in mind, that you have to export also dependencies, if their types are 
returned by (also already exported) packages. This is a limitation of gomobile and avoids cluttering 
your bindings with unwanted contracts. If an exported package is missing, `go get` will be invoked to grab it (e.g.
if it has not been pulled as a dependency already). Therefore it is valid to have only exports, and no imports.
However an empty export list makes no sense.

## hint
Important things like maps, slices, derived basic and value types won't work with gomobile. To mitigate these limitations, you can use the module https://github.com/worldiety/std. Do not forget to export it.
