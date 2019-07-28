# gollery
### (An attempt at replacing gallerpy)

Gollerpy is a photo gallery server which uses you folder structure to create the album on, it relies on json files (generated for your convenience) for extra metadata you might want to add and even allows you to hide files.
In simple words is a glorified file server, it uses a theme to render pages of the album and single pages, a barebones theme can be generated for you to extend.

I wrote this because i missed **gallerpy** that did something similar and was discontinued, you might serve this to the public but you will need an http server like nginx, apache or caddy to add https. I use this mainly for internal viewing of my own photos.
## Usage

```
Usage of gollery:
gollery [flags] /path/to/root/album:

--createmeta  (= false)
    create the metadata files for album and optionally subfolders
--createtheme (= "")
    create a theme for the album in the passed path with the passed name
--recursive  (= true)
    apply any action to the root album and subfolders
--theme (= "")
    use the passed in folder as a theme (included for creation)
```

## Install

```bash
go get -u github.com/perrito666/gollery
go install github.com/perrito666/gollery/...
```

## Configure

```bash
gollery --createmeta /path/to/root/album
# if you need a theme to modify (as opposed to have one)
gollery --createtheme --theme /path/to/theme /path/to/root/album
```

`--createmeta` creates `metadata.json` files in every album folder with the following structure:

```json
{
    "title": "",
    "pictures": {
        "_DSC1329.jpg": {
            "path": "playground\\_DSC1329.jpg",
            "file-name": "_DSC1329.jpg",
            "visible": true,
            "existing": true,
            "accessible": true
        },
        "_DSC1330.jpg": {
            "path": "playground\\_DSC1330.jpg",
            "file-name": "_DSC1330.jpg",
            "visible": true,
            "existing": true,
            "accessible": true
        },
    "order": [
        "_DSC1329.jpg",
        "_DSC1330.jpg",
    ],
    "sub-group-order": [
        "level1_1",
        "level1_2",
        "level1_3"
    ],
    "allowed-thumb-sizes": [
        {
            "width": 640,
            "height": 480
        }
    ]
}
```

The root element supports both `title` and `description` which are used for the folder page rendering (if not present folder name is used, at least in the default template).

`pictures` items support both `title` and `description` too, if you add them they will be displayed in the page for that image, also in the default template.

`allowed-thumb-sizes` is an array of items like the one above for the supported sizes of thumbnails in said folder, these will have effect only when you run `--createmeta` for creation and when a `GET` is made for said thumb (if not allowed you get a forbidden)

`--createtheme` creates a barebones theme skeleton for your to modify at leisure.

## Themes

The barebones theme has the following filesystem structure.

```
css/ 
html/  
js/
img/
templates/  
theme.json
```

The only mandatory items are in templates which are `page.html` and `single.html` which contain the template to render a folder and image respectively, you can use them as base or sample.

`theme.json` contains nothing of consequence now.

`css`, `html`, `js` and `img` are served under `/` as static file trees.

### Error files

If you create `404.html` and `500.html` files they will be used in the respective html errors (a more flexible support of errors to come).

