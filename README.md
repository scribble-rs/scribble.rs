<h1 align="center">Scribble.rs</h1>

<p align="center">
  <a href="https://github.com/scribble-rs/scribble.rs/actions">
    <img src="https://github.com/scribble-rs/scribble.rs/workflows/scribble-rs/badge.svg">
  </a>
  <a href="https://codecov.io/gh/scribble-rs/scribble.rs">
    <img src="https://codecov.io/gh/scribble-rs/scribble.rs/branch/master/graph/badge.svg">
  </a>
  <a href="https://discord.gg/3sntyCv">
    <img src="https://img.shields.io/discord/693433417395732531.svg?logo=discord">
  </a>
  <a href="https://heroku.com/deploy?template=https://github.com/scribble-rs/scribble.rs/tree/master">
    <img src="https://www.herokucdn.com/deploy/button.png">
  </a>
</p>

Scribble.rs is a clone of the web-based drawing game skribbl.io. In my opinion
skribbl.io has several usability issues, which I'll address in this project.

The site will not display any ads or share any data with third parties.

## Play now

Feel free to play on any of these instances:

* https://scribblers-official.herokuapp.com/
  > Might not respond right-away, just wait some seconds
* http://scribble.rs
  > No HTTPS!

### Hosting your own instance for free

By using Heroku, you can deploy a temporary container that runs scribble.rs.
The container will not have any cost and automatically suspend as soon as it
stops receiving traffic for a while.

Simply create an account at https://id.heroku.com/login and then click this link:

https://heroku.com/deploy?template=https://github.com/scribble-rs/scribble.rs/tree/master

## Building / Running

Run the following to build the application:

```shell
git clone https://github.com/scribble-rs/scribble.rs.git
cd scribble.rs
```

For -nix systems:
```shell
# run `make` to see all availables commands
make build
```

For Windows:
```shell
go run github.com/gobuffalo/packr/v2/packr2
go build -o scribblers .
```

This will produce a portable binary called `scribblers`. The binary doesn't
have any dependencies and should run on every system as long as it has the
same architecture and OS family as the system it was compiled on.

The default port will be `8080`. The parameter `portHTTP` allows changing the
port though.

It should run on any system that go supports as a compilation target.

This application uses go modules, therefore you need to make sure that you
have go version `1.13` or higher.

## Docker

Alternatively there's a docker container:

```shell
docker pull biosmarcel/scribble.rs
```

### Changing default port

The default port is `8080`. To override it, run:
```shell
docker run -p <port-number>:<port-number> biosmarcel/scribble.rs --portHTTP=<port-number>
```

## Contributing

There are many ways you can contribute:

* Update / Add documentation in the wiki of the GitHub repository
* Extend this README
* Create feature requests and bug reports
* Solve issues by creating Pull Requests
* Tell your friends about the project
* Curating the word lists

## Credits

These resources are by people unrelated to the project, whilst not every of these
resources requires attribution as per license, we'll do it either way ;)

If you happen to find a mistake here, please make a PR. If you are one of the
authors and feel like we've wronged you, please reach out.

* Favicon - [Fredy Sujono](https://www.iconfinder.com/freud)
* Rubber Icon - Made by [Pixel Buddha](https://www.flaticon.com/authors/pixel-buddha) from [flaticon.com](https://flaticon.com)
* Fill Bucket Icon - Made by [inipagistudio](https://www.flaticon.com/authors/inipagistudio) from [flaticon.com](https://flaticon.com)
* Kicking Icon - [Kicking Icon #309402](https://icon-library.net/icon/kicking-icon-4.html)
* Sound / No sound Icon - Made by Viktor Erikson (If this is you or you know who this is, send me a link to that persons Homepage)
* Profile Icon - Made by [kumakamu](https://www.iconfinder.com/kumakamu)
