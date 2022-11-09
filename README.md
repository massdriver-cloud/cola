# cola - CIDR Optimization by Lookup and Assignment

`COLA` is also the acronym for ["Collision On Launch Assessment"](https://nap.nationalacademies.org/read/13244/chapter/11#chapter09_r22), an assessment performed before a spacecraft launch to ensure there won't be a collision with another orbiting vehicle.

`cola` is a service for finding an available CIDR block given a parent CIDR block, a mask (desired block size), and list of already used CIDR blocks.

## Development

### Building

COLA is built using go:

```shell
go build
```

If you encounter an error: `fatal: could not read Username for 'https://github.com': terminal prompts disabled`, you can globally configure git:

```shell
git config --global --add url."git@github.com:".insteadOf "https://github.com/"
go build
```

### Adding Commands

Add commands using the [Cobra Generator](https://github.com/spf13/cobra/blob/master/cobra/README.md).

Commands should be scoped (subcommand) under a parent "command" to facilitate organization.

Blogs on writing Cobra commands:

* https://towardsdatascience.com/how-to-create-a-cli-in-golang-with-cobra-d729641c7177
