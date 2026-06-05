<!-- markdownlint-disable MD033 -->
<h1 align="center">plex2pl</h1>

<p align="center">
  <em>Convert your plex playlists into other formats for use with other programs.</em>
</p>

Supported formats:

* `jellyfin` native
* `m3u`

## Contents

* [Why](#why)
* [Install](#install)
  * [Download from GitHub](#download-from-github)
  * [Build it locally](#build-it-locally)
  * [Run the Docker container](#run-the-docker-container)
* [Configuring](#configuring)
* [Usage](#usage)
* [Running in Docker](#running-in-docker)
* [References](#references)

## Why

I used to use plexamp for listening to music, but have migrated all other media to Jellyfin.

## Install

### Download from GitHub

Find the latest version for your system on the
[GitHub releases page](https://github.com/tx3stn/plex2pl/releases).

### Build it locally

If you have go installed, you can clone this repo and run:

```bash
make install
```

This will build the binary and then copy it to `/usr/local/bin/plex2pl` so it will be
available on your path. Nothing more to it.

### Run the Docker container

Get the Docker container from the
[GitHub container registry](https://github.com/tx3stn/plex2pl/pkgs/container/plex2pl).

```bash
docker pull ghcr.io/tx3stn/plex2pl:latest
```

See [Running in Docker](#running-in-docker) for more details.

## Configuring

All of the configuration required for `plex2pl` is found in the config file.

The default expected locations for this are:

* `$XDG_CONFIG_DIR/plex2pl/config.json`
* `$HOME/.config/plex2pl/config.json`

If you want to use a file located somewhere else you can pass the `--config` flag, e.g.:

```bash
plex2pl --config /my/custom/config/file/path/config.json
```

> [!TIP]
> To get in editor feedback/validation of your schema, add the following to the top of your json file:
>
> ``` json
> "$schema": "https://raw.githubusercontent.com/tx3stn/plex2pl/refs/heads/main/.schema/schema.json"
> ```

### `plexServerUrl`

The url used to access your plex server, from the host device `plex2pl` is running on.

### `plexAuthToken`

The token required to authenticate the requests against the Plex server.

See [their docs on how you can find yours](https://support.plex.tv/articles/204059436).

### `outDirectory`

The location of the directory you want to generate the playlists in.

Each playlist will be created as a file with the playlist title as the name inside this directory.
Any path separator characters (`/` or `\`) in a playlist title are replaced with `-` in the generated file and directory names.

### `outputFormat`

The playlist format to generate. Supported values:

* `m3u` - `.m3u` files created directly inside `outDirectory`.
* `jellyfin` - jellyfin native playlists, created as `outDirectory/<playlist title>/playlist.xml` to match the layout of jellyfin's `data/playlists` directory.

The jellyfin format includes the genres of the tracks in the playlist.
If the genres are not returned in the playlist response from Plex, the track metadata is queried in a single batch request per playlist to resolve them.
If that request fails the playlist is still written, just without the missing genres.

### `jellyfinOwnerUserId`

The ID of the jellyfin user to set as the playlist owner when using the `jellyfin` output format.

Optional, when not set no owner is written to the playlist file.

You can find your user ID in the jellyfin admin dashboard under `Users`, it's the `userId` parameter in the URL when viewing a user's profile.

### `verbose`

If you want to see verbose output from the tool running.

Useful for debugging.

Can be enabled at run time (overriding the value in your config file), with the `--verbose` flag, e.g.:

```bash
plex2pl --verbose
```

## Usage

Once your config file is created, just run it:

```bash
plex2pl
```

That's it 🎉

The playlist files will be created in the directory you specified.

## Running in Docker

You can run this inside a container, with a few considerations:

1. **Volume mount your output directory**
So you can create files in the correct place, and not just inside the container.

2. **Volume mount your config**
A directory called `/config` is made for this, you can then pass the `--config` flag to the running command to use this config file.

3. **Specify the right user to run the container as**
So that your playlist files get created as the expected user to be able to use the created playlists, the example below uses the same user as the host.

4. **Make sure the container can access your plex server.**
If you're running plex in a container you will need to give access to the smae network plex is running on.

Putting this all together looks like this:

```bash
docker run --rm -v "/media/dir/music/playlists:/media/dir/music/playlists" \
 -v "/home/user/.config/plex2pl:/config" \
 --network host \
 -u $(id -u):$(id -g) \
 ghcr.io/tx3stn/plex2pl:latest --config "/config/config.json"
```

## References

* <https://github.com/XDGFX/PPP>
* <https://blog.fileformat.com/audio/common-errors-when-creating-or-editing-extm3u-files-and-how-to-fix-them/>
