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

First you'll need to install the Go compiler by followng the instructions at https://go.dev/doc/install.
If you are using a package manager for this, that's fine too.

Next you'll have to download the code via:

```shell
git clone https://github.com/scribble-rs/scribble.rs.git
cd scribble.rs
```

Lastly to build the executable, run the following:

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

You should be able to build the binary on any system that go supports as a compilation target.

This application requires go version `1.16` or higher.

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

## nginx 

Since Scribble.rs uses WebSockets, when running it behind an nginx reverse proxy, you have to configure nginx to support that.
You will find an example configuration on the [related Wiki page](https://github.com/scribble-rs/scribble.rs/wiki/reverse-proxy-(nginx)).

Other reverse proxies may require similar configuration. If you are using a well known reverse proxy, you are free to contribute a configuration to the wiki.

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

Some of these were slightly altered if the license allowed it.
Treat each of the files in this repository with the same license terms as the
original file.

* Logo - All rights reserved, excluded from BSD-3 licensing
* Favicon - All rights reserved, excluded from BSD-3 licensing
* Rubber Icon - Made by [Pixel Buddha](https://www.flaticon.com/authors/pixel-buddha) from [flaticon.com](https://flaticon.com)
* Fill Bucket Icon - Made by [inipagistudio](https://www.flaticon.com/authors/inipagistudio) from [flaticon.com](https://flaticon.com)
* Kicking Icon - [Kicking Icon #309402](https://icon-library.net/icon/kicking-icon-4.html)
* Sound / No sound Icon - Made by Viktor Erikson (If this is you or you know who this is, send me a link to that persons Homepage)
* Profile Icon - Made by [kumakamu](https://www.iconfinder.com/kumakamu)
* [Help Icon](https://www.iconfinder.com/icons/211675/help_icon) - Made by Ionicons
* [Fullscreen Icon](https://www.iconfinder.com/icons/298714/screen_full_icon) - Made by Github
* [Pencil Icon](https://github.com/twitter/twemoji/blob/8e58ae4/svg/270f.svg)
* [Checkmark Icon](https://commons.wikimedia.org/wiki/File:Green_check_icon_with_gradient.svg)
* [Fill Icon](https://commons.wikimedia.org/wiki/File:Circle-icons-paintcan.svg)
* [Trash Icon](https://www.iconfinder.com/icons/315225/trash_can_icon) - Made by [Yannick Lung](https://yannicklung.com)
* [Undo Icon](https://www.iconfinder.com/icons/308948/arrow_undo_icon) - Made by [Ivan Boyko](https://www.iconfinder.com/visualpharm)
