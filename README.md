# plex2m3u

Convert plex playlists into m3u format for use in other places e.g. Jellyfin

## Install

### Download from GitHub

Find the latest version for your system on the
[GitHub releases page](https://github.com/tx3stn/plex2m3u/releases).

### Build it locally

If you have go installed, you can clone this repo and run:

```bash
make install
```

This will build the binary and then copy it to `/usr/local/bin/plex2m3u` so it will be
available on your path. Nothing more to it.

### Run the Docker container

Get the Docker container from the
[GitHub container registry](https://github.com/tx3stn/plex2m3u/pkgs/container/plex2m3u).

```bash
docker pull ghcr.io/tx3stn/plex2m3u:latest
```

See [Running in Docker](#running-in-docker) for more details.

## Usage

1. Create your config file.
An example can be seen in [.schema/example.json](.schema/example.json)

This file tells the tool how to work, if it's not found then `plex2m3u` can't do anything.

The default expected locations for this are:
* `$XDG_CONFIG_DIR/plex2m3u/config.json`
* `$HOME/.config/plex2m3u/config.json`

If you want to use a file located somewhere else you can pass the `--config` flag, e.g.:

```bash
plex2m3u --config /my/custom/config/file/path/config.json
```

> [!NOTE]
> To get the data from Plex, you need an auth token.
>
> See [their docs on how you can find yours](https://support.plex.tv/articles/204059436).

2. Run it!
That's it, once your config file exists you can just run `plex2m3u` and it will get all of your
audio playlists and create an `m3u` file for them in the specified output path.

If you want to see more information about what's happening during a run, you can enable verbose output with the `--verbose` flag, e.g.:

```bash
plex2m3u --verbose
```

## Running in Docker

You can run this inside a container, with a few considerations:
1. Volume mount your output directory
So you can create files in the correct place, and not just inside the container.
2. Volume mount your config
A directory called `/config` is made for this, you can then pass the `--config` flag to the running command to use this config file.
3. Specify the right user to run the container as
So that your playlist files get created as the expected user to be able to use the created playlists, the example below uses the same user as the host.
4. Make sure the container can access your plex server.
If you're running plex in a container you will need to give access to the smae network plex is running on.

Putting this all together looks like this:

```bash
docker run --rm -v "/media/dir/music/playlists:/media/dir/music/playlists" \
	-v "/home/user/.config/plex2mu3:/config" \
	--network host \
	-u $(id -u):$(id -g) \
	plex2m3u:local --config "/config/config.json"
```

## References

- https://github.com/XDGFX/PPP
- https://blog.fileformat.com/audio/common-errors-when-creating-or-editing-extm3u-files-and-how-to-fix-them/
