Overview
========

What is Khan? Khan is an HTTP "resty" API for managing clans for games. It could be used to manage groups of people, but our aim is players in a game.

Khan allows your app to focus on the interaction required to creating clans and managing applications, instead of the backend required for actually doing it.

## Features

* **Multi-tenant** - Khan already works for as many games as you need, just keep adding new games;
* **Clan Management** - Create and manage clans, their metadata as well as promote and demote people in their rosters;
* **Player Management** - Manage players and their metadata, as well as their applications to clans;
* **Applications** - Khan handles the work involved with applying to clans, inviting people to clans, accepting, denying and kicking;
* **Clan Search** - Search a list of clans to present your player with relevant options;
* **Top Clans** - Choose from a specific dimension to return a list of the top clans in that specific range (SOON);
* **Web Hooks** - Need to integrate your clan system with another application? We got your back! Use our web hooks sytem and plug into whatever events you need;
* **Auditing Trail** - Track every action coming from your games (SOON);
* **Easy to deploy** - Khan comes with containers already exported to docker hub for every single of our successful builds. Just pick your choice!

## Architecture

Khan is based on the premise that you have a backend server for your game. That means we do not employ any means of authentication.

There's no validation if the actions you are performing are valid as well. We have TONS of validation around the operations themselves being valid.

What we don't have are validations that test whether the source of the request can perform the request (remember the authentication bit?).

Khan also offers a JSON `metadata` field in its Player and Clan models. This means that your game can store relevant information that can later be used to sort players, for example.

## The Stack

For the devs out there, our code is in Go, but more specifically:

* Web Framework - [Echo](https://github.com/labstack/echo) based on the insanely fast [FastHTTP](https://github.com/valyala/fasthttp);
* Database - Postgres >= 9.5;
* Cache - Redis.

## Who's Using it

Well, right now, only us at TFG Co, are using it, but it would be great to get a community around the project. Hope to hear from you guys soon!

## How To Contribute?

Just the usual: Fork, Hack, Pull Request. Rinse and Repeat. Also don't forget to include tests and docs (we are very fond of both).
