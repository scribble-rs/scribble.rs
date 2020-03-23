# Scribble.rs

[![CircleCI](https://circleci.com/gh/scribble-rs/scribble.rs.svg?style=svg)](https://circleci.com/gh/scribble-rs/scribble.rs)
[![codecov](https://codecov.io/gh/scribble-rs/scribble.rs/branch/master/graph/badge.svg)](https://codecov.io/gh/scribble-rs/scribble.rs)

This project is intended to be a clone of the web-based drawing game
[skribbl.io](https://skribbl.io), since skribbl.io has several usuability
issues. This project will try addressing as many of the usuability issues
as possible.

Even though there is an official instance at
[scribble.rs](http://scribble.rs), you can still host your own instance.

The site will not display any ads or share any data with third parties.

## Building

Run the following to build the application:

```shell
git clone https://github.com/scribble-rs/scribble.rs.git
cd scribble.rs
go build -o scribblers .
```

This will produce a binary called `scribblers`. The binary will still depend
on the sourec folder and should just be called from within. This is due to the
fact, that the HTML templates and the resource data aren't part of the binary.

The default port will be `8080`. The parameter `portHTTP` allows changing the
port though.

It should run on any system that go supports as a compilation target.

This application uses go modules, therefore you need to make sure that you
have go version `1.11` or higher and the environment variable `GO111MODULE`
set to `on`.

## Contributing

There are many ways you can contribute:

* Update / Add documentation in the wiki of the GitHub repository
* Extend this README
* Create issues
* Solve issues by creating Pull Requests
* Tell your friends about the project
* Curating the word lists

## Credits

* Favicon - [Fredy Sujono](https://www.iconfinder.com/freud)
* Rubber Icon - Made by [Pixel Buddha](https://www.flaticon.com/authors/pixel-buddha) from [flaticon.com](https://flaticon.com)
* Fill Bucket Icon - Made by [inipagistudio](https://www.flaticon.com/authors/inipagistudio) from [flaticon.com](https://flaticon.com)
