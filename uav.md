# Using `uav`
If the readme is the "why" then this is the "how".

## Requirements
1. c++14 capable compiler (any recent `clang` or `gcc` will do!)
1. yaml-cpp

## Building `uav`
In this directory, run `make`:

```bash
$ make
```

to build `uav`. Outputs are placed in the `bin` folder. By default, `uav` is made as a statically linked library. `uav` has been sucessfully built on macOS 10.10, and on Ubuntu 18.04. There should be no special considerations for Windows builds, although this technically this hasn't been tested..

There are some tests that can be ran, this is acheived with:

```bash
$ make tests
```

## Using `uav`
`uav` does not invoke `fly`, rather, it just generates the output pipeline file. If you wish to use a pipeline definition generated with `uav`, you'll still need to `fly` it as normal.

If you don't use any features of `uav`, the pipeline should be passed through untouched. Semantically, at least - the output of `uav` is JSON. We're not at all bothered by the representation of the pipeline, only C++ library support for JSON is preferable to YAML.

If you have a pipeline you'd like to try out, say `pipeline.yaml`, try running this:

```bash
$ bin/uav pipeline.yaml
```

Hopefully you'll get the same pipeline out as in. The default output path will be `pipeline.json`. You can control the output path of `uav` with the `-o` or `--output` argument:

```bash
$ bin/uav pipeline.yaml --output p.json
```

## Credentials
As was mentioned in the [readme](./readme.md), variables must be substituted at generation time. Local population of credentials uses these types of variables - so they must be supplied at generation time. This is achieved with the `-c` or `--credentials` option:

```bash
$ bin/uav pipeline.yaml --credentials credentials.yaml
```

## Command line variables
You can supply multiple arguments to a pipeline templating run:

```bash
$ bin/uav pipeline.yml -d "product=product-1" -d "configuration=Release"
```

It is acceptable for the value of a substitution to be empty, or contain additional `=` characters.

## Examples
An example is inluded in the `examples` folder. Once uav has been built, you can run it with:

```bash
$ bin/uav examples/example-1/pipeline.yml
```

This will produce `pipeline.json` - with all of the pieces templated into the one file. In principle, this would be the pipeline that `fly` would send to ATC.
