# hoi4treesnap

hoi4treesnap generates Hearts of Iron IV focus tree screenshots.

The tool itself does not contain any textures and picks them up from the HOI4 base game or a mod that contains selected focus trees. That includes all focus tree graphics: focus icons, focus tree plaques, focus tree lines and fonts. `nationalfocusview.gui` is being parsed to pick on your changes to it, so the output image looks quite similar to what you see in the game, even a modded one.

## Linux support

hoi4treesnap now builds and runs natively on Debian Linux amd64 with Fyne v2.

Install the required development packages before building:

```bash
sudo apt-get update
sudo apt-get install -y \
  libgl1-mesa-dev \
  libx11-dev \
  libxrandr-dev \
  libxxf86vm-dev \
  libxi-dev \
  libxcursor-dev \
  libxinerama-dev
```

Build the portable Linux binary with:

```bash
make build-linux
```

This produces `dist/hoi4treesnap-linux-amd64`.

Fully static linking with `-ldflags="-extldflags=-static"` is not practical for the Fyne/GLFW/OpenGL stack on glibc-based Debian systems, so the Linux binary remains dynamically linked. At runtime it still relies on the usual system libraries from the OpenGL/X11 stack, including `libGL.so.1`, `libX11.so.6`, `libXrandr.so.2`, `libXxf86vm.so.1`, `libXi.so.6`, `libXcursor.so.1`, `libXinerama.so.1`, and glibc.

The saved HOI4 path cache is now stored under `os.UserCacheDir()`, which is typically `~/.cache/hoi4treesnap/hoi4treesnapGamePath.txt` on Linux and `%LocalAppData%\hoi4treesnap\hoi4treesnapGamePath.txt` on Windows.

Linux builds auto-detect the base game at these locations before falling back to the saved cache:

1. `~/.steam/steam/steamapps/common/Hearts of Iron IV`
2. `~/.var/app/com.valvesoftware.Steam/data/Steam/steamapps/common/Hearts of Iron IV`

## How to use

1. Download and run the latest binary from [the releases page](https://github.com/malashin/hoi4treesnap/releases).
2. Select focus tree file from `/common/national_focus`.
3. Select Hearts of Iron IV game folder. It will be saved for later use after the first time if auto-detection did not already find it.
4. If you need other mods, dependencies for example, select those.
5. If you want to use non-english localisation press `Select localisation language`.
6. Press `Generate image`. Output will be saved next to the hoi4treesnap binary.

## DDS coverage

The DDS decoder is now backed by `github.com/xypwn/filediver/dds`, which registers with the standard Go image package and correctly handles the cases that matter for HOI4 and mod textures:

1. DXT1 and DXT5 textures even when `pitchOrLinearSize` is zero or incorrect.
2. Mipmapped textures while decoding only the base image for normal rendering.
3. DX10 extended headers.
4. Uncompressed 32-bit ARGB8 textures with correct BGRA to RGBA channel handling.

The test suite includes synthetic DDS fixtures covering DXT1, DXT5 with alpha, uncompressed ARGB8, and a DX10 mipmapped texture.

## CI and releases

GitHub Actions now builds Linux and Windows binaries on pushes to `master` and on version tags matching `v*`. Tag builds also publish both binaries as release assets. The Windows build runs on a native `windows-latest` runner because this Fyne/go-gl stack is not a reliable Linux-to-Windows cross-compilation target.

## Possible issues

* The file parser is stricter then PDX one, so you might need to fix those errors if they are reported.

## Known issues

* You can't generate single image for shared focus trees. You'll have to combine them from separate images.
* There is no country name in the image. Might be added later either through parsing of the files or just asking the user to input the name.
* If focus title uses scripted localization, it will be rendered as a scripted localization string instead of the appropriate name. Might ask user to enter appropriate titles if those are found later on.

## Menu

![TreeSnap menu](https://i.imgur.com/84sotcl.png)

## Output examples

![Output example 1](https://i.imgur.com/MKPV5Cc.png)
![Output example 2](https://i.imgur.com/8Bq71l1.png)
