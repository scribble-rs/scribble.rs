<h1 align="center">Scribble.rs</h1>

<p align="center">
  <a href="https://github.com/scribble-rs/scribble.rs/actions">
    <img src="https://github.com/scribble-rs/scribble.rs/workflows/Run%20scribble-rs%20tests/badge.svg">
  </a>
  <a href="https://codecov.io/gh/scribble-rs/scribble.rs">
    <img src="https://codecov.io/gh/scribble-rs/scribble.rs/branch/master/graph/badge.svg">
  </a>
  <a href="https://liberapay.com/biosmarcel/donate">
    <img src="https://img.shields.io/liberapay/receives/biosmarcel.svg?logo=liberapay">
  </a>
  <a href="https://heroku.com/deploy?template=https://github.com/scribble-rs/scribble.rs/tree/master">
    <img src="https://www.herokucdn.com/deploy/button.png">
  </a>
</p>

Scribble.rs is an alternative to the web-based drawing game skribbl.io. My main
problems with skribbl.io were the ads and the fact that a disconnect would
cause you to lose your points. On top of that, the automatic word choice was
quite annoying and caused some frustration.

The site will not display any ads or share any data with third parties.

## News and discussion

We have a new blog over at https://scribble-rs.github.io. Over there, you can read about some highlights and discuss them in the comment section.
The comment section is powered by utteranc.es, which means it'll use the blogs repository for comments and you can simply use your GitHub account for commenting.

## Play now

Feel free to play on this instance

* https://scribblers-official.herokuapp.com/
  > Might not respond right-away, just wait some seconds / minutes, as it
  > shuts down automatically when unused.

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
# run `make` to see all available commands
make build
```

For Windows:
```shell
go build -o scribblers .
```

This will produce a portable binary called `scribblers`. The binary doesn't
have any dependencies and should run on every system as long as it has the
same architecture and OS family as the system it was compiled on.

The default port will be `8080`. The parameter `portHTTP` allows changing the
port though.

It should run on any system that go supports as a compilation target.

This application uses go modules, therefore you need to make sure that you
have go version `1.16` or higher.

## Docker

Alternatively there's a docker container:

```shell
docker pull biosmarcel/scribble.rs
```

The docker container is built from the master branch on every push, so it
should always be up-to-date.

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

## Donating

If you can't or don't want to contribute in any of the ways
listed above, you can always donate something to the project.

* PayPal: https://www.paypal.com/donate/?hosted_button_id=RZ7N8D95TXFEN
* Liberapay: https://liberapay.com/biosmarcel/donate
* Etherum: 0x49939106563a9de8a777Cf5394149423b1dFd970
* XLM/Lumen: GDNCEW46OTDMXMSNVM4K7GNPIXNYT5BOZXVZ7M4QSRB6OB3BRM2VYDB5

If there's a steady income stream I'd spend it on infrastructure and a domain ;)

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
