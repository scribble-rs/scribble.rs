<h1 align="center">Scribble.rs</h1>

<p align="center">
  <a href="https://circleci.com/gh/scribble-rs/scribble.rs">
    <img src="https://circleci.com/gh/scribble-rs/scribble.rs.svg?style=svg">
  </a>
  <a href="https://codecov.io/gh/scribble-rs/scribble.rs">
    <img src="https://codecov.io/gh/scribble-rs/scribble.rs/branch/master/graph/badge.svg">
  </a>
  <a href="https://discord.gg/3sntyCv">
    <img src="https://img.shields.io/discord/693433417395732531.svg?label=&logo=discord&logoColor=ffffff&color=7389D8&labelColor=6A7EC2">
  </a>
</p>

This project is intended to be a clone of the web-based drawing game
[skribbl.io](https://skribbl.io). In my opinion skribbl.io has several
usability issues, which I'll try addressing in this project.

Even though there is an official instance at
[scribble.rs](http://scribble.rs), you can still host your own instance.

The site will not display any ads or share any data with third parties.

## Building / Running

Run the following to build the application:

```shell
git clone https://github.com/scribble-rs/scribble.rs.git
cd scribble.rs
go get github.com/markbates/pkger/cmd/pkger
pkger -include /resources -include /templates
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
