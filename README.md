# discord-gif-raster
### Command line program to split a gif into blocks to be uploaded and boasted about in discord
---
To use it:
* Get the binary (probably by downloading it from releases)
* I recommend you place in a folder with the gif you want to rasterize
* Run it with `raster -f <filename.gif>`
* Upload newly generated files in gifs/ directory (you can multi-upload if you select all on discord)
* Copy the text and paste into your favorite server

To build:
* `go get bitbucket.org/marvinody/discord-gif-raster`
* `go build -o raster bitbucket.org/marvinody/discord-gif-raster`

Sample:
![bunpats.gif](bunpats.gif)
![000.gif](gifs/000.gif) ![001.gif](gifs/001.gif)
![002.gif](gifs/002.gif) ![003.gif](gifs/003.gif)
![004.gif](gifs/004.gif) ![005.gif](gifs/005.gif)
