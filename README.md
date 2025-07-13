# plex2pl

Convert plex playlists into m3u format for use in other places e.g. Jellyfin

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
> ``` json
> "$schema": "https://raw.githubusercontent.com/tx3stn/plex2pl/refs/heads/main/.schema/schema.json"
> ```

### `plexServerUrl`

The url used to access your plex server, from the host device `plex2pl` is running on.

### `plexAuthToken`

The token required to authenticate the requests against the Plex server.

See [their docs on how you can find yours](https://support.plex.tv/articles/204059436).

### `OutDirectory`

The location of the directory you want to generate the playlists in.

Each playlist will be created as a file with the playlist title as the name inside this directory.

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
	-v "/home/user/.config/plex2mu3:/config" \
	--network host \
	-u $(id -u):$(id -g) \
	plex2pl:local --config "/config/config.json"
```

## References

- https://github.com/XDGFX/PPP
- https://blog.fileformat.com/audio/common-errors-when-creating-or-editing-extm3u-files-and-how-to-fix-them/
