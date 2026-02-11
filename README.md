# faketunes

A proof-of-concept virtual storage with ALAC music files as a FUSE filesystem.

## What it does?

Let's assume you have a music library of FLACs with different sample rates and bitrates. This folder is having the following structure:

```
/library/Artist/Album/01 - Track Name.flac
```

(this is the default music library structure for [beets](https://beets.io/)).

Let's also assume you have an iPod classic with the original firmware and you want to sync your library to it. But you also don't want to convert all your huge library into other directory for iTunes to consume, taking twice as much space for your music library collection and duplicating files.

This is when `faketunes` comes handy. It makes a virtual FUSE filesystem that represents all your FLACs as ALACs into a single folder called `Music` inside a path you chose. You can then mount that folder as a network share and point iTunes on Mac or Windows to use it as a music library folder. After that, you can add all the files into the iTunes library, and in the end, sync your iPod Classic.

There are no actual ALAC files in the `Music` directory: they're all virtual. On the first attempt to access the file, `faketunes` will generate with `ffmpeg` an ALAC file with proper metadata (taken from your source FLAC files) and album art, place it in the cache and serve it. You can tune the cache size in the config (see below). All subsequental reads for the file will be provided from cache as long as the converted file is present: otherwise, `ffmpeg` will be run again.

The goals of the project:

- Make a virtual filesystem to serve files to iTunes and to not convert the FLAC music library for iPod Classic separately
- Track the original library and make sure there are no missing unconverted files.
- Reduce hard drive space usage (HDDs are not cheap anymore because of AI...)
- Be as fast and reliable as possible.
- To learn something new about file systems in general.

## Status of the project

Currently, the project is in _alpha_ state. It (somewhat) works, and serves as a proof that making such a virtual filesystem is possible in the first place. Optimizations and patches are welcome.

## Requirements

- Linux host (Docker image will come later)
- FUSE support (and `fusermount3` command present in the `$PATH`)
- `ffmpeg` and `ffprobe` installed on the system.

## Configuration

By default, `faketunes` searches for config at the `/etc/faketunes.yml`. You can override the config path by providing the environment variable `FAKETUNES_CONFIG` with the desired path.

See `faketunes.example.yaml` file in the repo for the configuration example.
